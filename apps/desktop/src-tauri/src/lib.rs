use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use tauri::Manager;
use tauri::State;

use std::path::{Path, PathBuf};
use std::process::Child;
use std::sync::Mutex;

static BACKEND_PROCESS: Mutex<Option<Child>> = Mutex::new(None);

struct ApiToken(String);

#[tauri::command]
fn get_api_token(token: State<'_, ApiToken>) -> String {
    token.0.clone()
}

#[tauri::command]
fn relaunch(app_handle: tauri::AppHandle) {
    app_handle.restart();
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let mut token = String::new();

    // Dev/Prod check: Try to read existing token from file first so we sync with separately run backend
    let app_dir = get_config_dir();
    let local_token_file = app_dir.join(".auth_token");

    // Also support dev path check when running in the dev repo folder
    let mut dev_token_file = std::path::PathBuf::new();
    if let Ok(path) = std::env::current_dir() {
        dev_token_file = path.join("../backend/data/.auth_token");
    }

    if !dev_token_file.as_os_str().is_empty() {
        if let Ok(data) = std::fs::read_to_string(&dev_token_file) {
            token = data.trim().to_string();
        }
    }

    if token.is_empty() {
        if let Ok(data) = std::fs::read_to_string(&local_token_file) {
            token = data.trim().to_string();
        }
    }

    // If no existing token, generate a new one
    if token.is_empty() || token.len() < 16 {
        token = thread_rng()
            .sample_iter(&Alphanumeric)
            .take(32)
            .map(char::from)
            .collect();
    }

    // Set environment variable for any spawned subprocesses
    std::env::set_var("RAYFLOW_AUTH_TOKEN", &token);

    // Save token to all locations (development backend and production app configs)
    let _ = std::fs::create_dir_all(&app_dir);
    let _ = std::fs::write(local_token_file, &token);

    if let Ok(path) = std::env::current_dir() {
        let dev_dir = path.join("../backend/data");
        if dev_dir.exists() {
            let _ = std::fs::write(dev_dir.join(".auth_token"), &token);
        }
    }

    let app = tauri::Builder::default()
        .plugin(tauri_plugin_updater::Builder::new().build())
        .manage(ApiToken(token.clone()))
        .invoke_handler(tauri::generate_handler![get_api_token, relaunch])
        .setup(move |app| {
            if let Some(path) = find_sidecar_path(app) {
                if let Ok(child) = std::process::Command::new(path)
                    .env("RAYFLOW_AUTH_TOKEN", &token)
                    .spawn()
                {
                    let mut lock = BACKEND_PROCESS.lock().unwrap();
                    *lock = Some(child);
                }
            }

            Ok(())
        })
        .build(tauri::generate_context!())
        .expect("error while running RayFlow Client");

    app.run(|_app_handle, event| match event {
        tauri::RunEvent::Exit => {
            // Clean up the backend process when the frontend window exits
            let mut lock = BACKEND_PROCESS.lock().unwrap();
            if let Some(mut child) = lock.take() {
                let _ = child.kill();
            }
        }
        _ => {}
    });
}

fn get_config_dir() -> std::path::PathBuf {
    if let Ok(appdata) = std::env::var("APPDATA") {
        std::path::PathBuf::from(appdata).join("RayFlow")
    } else if let Ok(home) = std::env::var("HOME") {
        std::path::PathBuf::from(home).join(".config").join("RayFlow")
    } else {
        std::path::PathBuf::from("data")
    }
}

fn find_sidecar_path(app: &tauri::App) -> Option<PathBuf> {
    let mut dirs: Vec<PathBuf> = Vec::new();

    if let Ok(resource_dir) = app.path().resource_dir() {
        dirs.push(resource_dir.clone());
        dirs.push(resource_dir.join("bin"));
    }

    if let Ok(exe_path) = std::env::current_exe() {
        if let Some(exe_dir) = exe_path.parent() {
            dirs.push(exe_dir.to_path_buf());
            dirs.push(exe_dir.join("bin"));
            dirs.push(exe_dir.join("../../../bin"));
        }
    }

    if let Ok(current_dir) = std::env::current_dir() {
        dirs.push(current_dir.join("apps/desktop/src-tauri/bin"));
        dirs.push(current_dir.join("src-tauri/bin"));
        dirs.push(current_dir.join("bin"));
    }

    for dir in dirs {
        if let Some(path) = first_existing_sidecar(&dir) {
            return Some(path);
        }
    }

    None
}

fn first_existing_sidecar(dir: &Path) -> Option<PathBuf> {
    for name in sidecar_file_names() {
        let path = dir.join(name);
        if path.exists() {
            return Some(path);
        }
    }
    None
}

fn sidecar_file_names() -> &'static [&'static str] {
    #[cfg(target_os = "windows")]
    {
        &["rayflowd.exe", "rayflowd-x86_64-pc-windows-msvc.exe"]
    }

    #[cfg(target_os = "macos")]
    {
        if cfg!(target_arch = "aarch64") {
            &["rayflowd", "rayflowd-aarch64-apple-darwin"]
        } else {
            &["rayflowd", "rayflowd-x86_64-apple-darwin"]
        }
    }

    #[cfg(target_os = "linux")]
    {
        &["rayflowd", "rayflowd-x86_64-unknown-linux-gnu"]
    }
}
