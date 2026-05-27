use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use tauri::State;

struct ApiToken(String);

#[tauri::command]
fn get_api_token(token: State<'_, ApiToken>) -> String {
    token.0.clone()
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let token: String = thread_rng()
        .sample_iter(&Alphanumeric)
        .take(32)
        .map(char::from)
        .collect();

    // Set environment variable for any spawned subprocesses
    std::env::set_var("RAYFLOW_AUTH_TOKEN", &token);

    // Save token to file for backend process reading (dev and prod fallbacks)
    if let Ok(path) = std::env::current_dir() {
        // Dev fallback: apps/backend/data
        let dev_dir = path.join("../backend/data");
        if dev_dir.exists() {
            let _ = std::fs::write(dev_dir.join(".auth_token"), &token);
        }
        // Prod / Local fallback: data/
        let local_dir = path.join("data");
        let _ = std::fs::create_dir_all(&local_dir);
        let _ = std::fs::write(local_dir.join(".auth_token"), &token);
    }

    tauri::Builder::default()
        .manage(ApiToken(token))
        .invoke_handler(tauri::generate_handler![get_api_token])
        .run(tauri::generate_context!())
        .expect("error while running RayFlow Client");
}

