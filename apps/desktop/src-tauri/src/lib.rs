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
    let mut token = String::new();

    // Dev/Prod check: Try to read existing token from file first so we sync with separately run backend
    if let Ok(path) = std::env::current_dir() {
        let dev_token_file = path.join("../backend/data/.auth_token");
        let local_token_file = path.join("data/.auth_token");

        if let Ok(data) = std::fs::read_to_string(&dev_token_file) {
            token = data.trim().to_string();
        } else if let Ok(data) = std::fs::read_to_string(&local_token_file) {
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

