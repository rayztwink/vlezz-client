ALTER TABLE app_settings ADD COLUMN default_core TEXT NOT NULL DEFAULT 'sing-box';
ALTER TABLE app_settings ADD COLUMN local_proxy_port INTEGER NOT NULL DEFAULT 2080;
ALTER TABLE app_settings ADD COLUMN sing_box_path TEXT NOT NULL DEFAULT '';
ALTER TABLE app_settings ADD COLUMN xray_path TEXT NOT NULL DEFAULT '';
ALTER TABLE app_settings ADD COLUMN zapret_path TEXT NOT NULL DEFAULT '';

