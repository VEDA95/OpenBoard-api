CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "open_board_user" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "external_provider_id" UUID,
    "thumbnail" UUID,
    "username" VARCHAR(255) UNIQUE NOT NULL,
    "email" VARCHAR(255) UNIQUE NOT NULL,
    "first_name" VARCHAR(255),
    "last_name" VARCHAR(255),
    "dark_mode" BOOLEAN NOT NULL DEFAULT (false),
    "hashed_password" TEXT,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "last_login" TIMESTAMP,
    "enabled" BOOLEAN NOT NULL DEFAULT (true),
    "email_verified" BOOLEAN NOT NULL DEFAULT (false),
    "reset_password_on_login" BOOLEAN NOT NULL DEFAULT (false)
);

CREATE TABLE "open_board_user_session" (
   "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
   "user_id" UUID NOT NULL,
   "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
   "date_updated" TIMESTAMP,
   "expires_on" TIMESTAMP NOT NULL,
   "session_type" VARCHAR(32) NOT NULL,
   "remember_me" BOOLEAN NOT NULL DEFAULT (false),
   "access_token" TEXT UNIQUE,
   "refresh_token" TEXT UNIQUE,
   "ip_address" VARCHAR(255) NOT NULL,
   "user_agent" VARCHAR(255) NOT NULL,
   "additional_info" JSONB
);

CREATE TABLE "open_board_external_auth_provider" (
     "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
     "name" VARCHAR(255) NOT NULL,
     "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
     "date_updated" TIMESTAMP,
     "client_id" VARCHAR(255) NOT NULL,
     "client_secret" TEXT,
     "use_pkce" BOOLEAN NOT NULL DEFAULT (false),
     "auth_url" VARCHAR(255) NOT NULL,
     "token_url" VARCHAR(255) NOT NULL,
     "userinfo_url" VARCHAR(255) NOT NULL,
     "logout_url" VARCHAR(255),
     "default_login_method" BOOLEAN NOT NULL DEFAULT (false),
     "self_registration_enabled" BOOLEAN NOT NULL DEFAULT (false),
     "required_email_domain" VARCHAR(255)
);

CREATE TABLE "open_board_role" (
   "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
   "name" VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE "open_board_role_permission" (
  "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
  "path" VARCHAR(255) NOT NULL
);

CREATE TABLE "open_board_multi_auth_challenge" (
    "id" TEXT PRIMARY KEY,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "expires_on" TIMESTAMP NOT NULL,
    "auth_method_id" UUID,
    "user_id" UUID NOT NULL,
    "data" JSONB
);

CREATE TABLE "open_board_multi_auth_method" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "user_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "name" VARCHAR(255) NOT NULL,
    type VARCHAR(16) NOT NULL,
    "credential_data" JSONB NOT NULL
);

CREATE TABLE "open_board_user_password_reset_token" (
    "id" TEXT PRIMARY KEY,
    "user_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "expires_on" TIMESTAMP NOT NULL
);

CREATE TABLE "open_board_user_email_verification_token" (
    "id" TEXT PRIMARY KEY,
    "user_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "expires_on" TIMESTAMP NOT NULL
);

CREATE TABLE "open_board_general_settings" (
   "app_name" VARCHAR(255),
   "logo" UUID,
   "site_url" TEXT
);

CREATE TABLE "open_board_notification_settings" (
    "smtp_server" VARCHAR(255),
    "smtp_port" INTEGER,
    "smtp_user" VARCHAR(255),
    "smtp_password" TEXT,
    "name" VARCHAR(255),
    "email_address" VARCHAR(255),
    "use_tls" BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE "open_board_auth_settings" (
    "access_token_lifetime" INTEGER NOT NULL DEFAULT 3600,
    "refresh_token_lifetime" INTEGER NOT NULL DEFAULT 7200,
    "refresh_token_idle_lifetime" INTEGER NOT NULL DEFAULT 1209600,
    "multi_factor_auth_enabled" BOOLEAN NOT NULL DEFAULT TRUE,
    "force_multi_factor_auth" BOOLEAN NOT NULL DEFAULT FALSE,
    "otp_enabled" BOOLEAN NOT NULL DEFAULT TRUE,
    "authenticator_enabled" BOOLEAN NOT NULL DEFAULT TRUE,
    "webauthn_enabled" BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE "open_board_file_upload" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "user_id" UUID NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "file_path" VARCHAR(255) NOT NULL,
    "file_size" INTEGER NOT NULL,
    "additional_details" JSONB
);

CREATE TABLE "open_board_workspace" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT
);

CREATE TABLE "open_board_board" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "user_id" UUID NOT NULL,
    "workspace_id" UUID,
    "name" VARCHAR(255) NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "is_public" BOOLEAN NOT NULL DEFAULT (true)
);

CREATE TABLE "open_board_board_label" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "board_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "name" VARCHAR(255) NOT NULL,
    "color" VARCHAR(7)
);

CREATE TABLE "open_board_board_field" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "board_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "name" VARCHAR(255) NOT NULL,
    "type" VARCHAR(32) NOT NULL,
    "field_properties" JSONB NOT NULL
);

CREATE TABLE "open_board_board_list" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "board_id" UUID NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "color" VARCHAR(7),
    "position" INTEGER NOT NULL
);

CREATE TABLE "open_board_board_list_card" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "list_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "reminder_date" TIMESTAMP,
    "due_date" TIMESTAMP,
    "name" VARCHAR(255) NOT NULL,
    "priority" INTEGER,
    "time_spent" INTEGER,
    "estimated_time_spent" INTEGER,
    "description" TEXT,
    "is_active" BOOLEAN NOT NULL DEFAULT (true)
);

CREATE TABLE "open_board_board_list_card_comment" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "user_id" UUID NOT NULL,
    "card_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "comment" TEXT NOT NULL
);

CREATE TABLE "open_board_board_list_card_checklist_item" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "card_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "date_updated" TIMESTAMP,
    "name" VARCHAR(255) NOT NULL,
    "is_checked" BOOLEAN NOT NULL DEFAULT (false),
    "position" INTEGER NOT NULL
);

CREATE TABLE "open_board_board_list_card_card_activity" (
    "id" UUID PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "card_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "date_created" TIMESTAMP NOT NULL DEFAULT (now()),
    "activity" VARCHAR(255) NOT NULL
);

CREATE TABLE "open_board_role_permissions" (
    "role_id" UUID NOT NULL,
    "permission_id" UUID NOT NULL,
    PRIMARY KEY ("role_id", "permission_id")
);

CREATE TABLE "open_board_user_roles" (
    "user_id" UUID NOT NULL,
    "role_id" UUID NOT NULL,
    PRIMARY KEY ("user_id", "role_id")
);

CREATE TABLE "open_board_external_provider_roles" (
    "provider_id" UUID NOT NULL,
    "role_id" UUID NOT NULL,
    PRIMARY KEY ("provider_id", "role_id")
);

CREATE TABLE "open_board_workspace_permissions" (
    "workspace_id" UUID NOT NULL,
    "permission_id" UUID NOT NULL,
    PRIMARY KEY ("workspace_id", "permission_id")
);

CREATE TABLE "open_board_board_permissions" (
    "board_id" UUID NOT NULL,
    "permission_id" UUID NOT NULL,
    PRIMARY KEY ("board_id", "permission_id")
);

CREATE TABLE "open_board_workspace_admins" (
   "workspace_id" UUID NOT NULL,
   "user_id" UUID NOT NULL,
    PRIMARY KEY ("workspace_id", "user_id")
);

CREATE TABLE "open_board_workspace_members" (
    "workspace_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    PRIMARY KEY ("workspace_id", "user_id")
);

CREATE TABLE "open_board_board_list_card_attachments" (
    "card_id" UUID NOT NULL,
    "file_id" UUID NOT NULL,
    PRIMARY KEY ("card_id", "file_id")
);

CREATE TABLE "open_board_board_list_card_fields" (
    "field_id" UUID NOT NULL,
    "card_id" UUID NOT NULL,
    "position" INTEGER NOT NULL,
    PRIMARY KEY ("field_id", "card_id")
);

CREATE TABLE "open_board_board_list_card_labels" (
    "label_id" UUID NOT NULL,
    "card_id" UUID NOT NULL,
    PRIMARY KEY ("label_id", "card_id")
);

CREATE TABLE "open_board_board_members" (
    "board_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    PRIMARY KEY ("board_id", "user_id")
);

CREATE TABLE "open_board_board_list_card_members" (
    "card_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    PRIMARY KEY ("card_id", "user_id")
);

ALTER TABLE "open_board_user_session" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_user_password_reset_token" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_user_email_verification_token" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_multi_auth_method" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_multi_auth_challenge" ADD FOREIGN KEY ("auth_method_id") REFERENCES "open_board_multi_auth_method" ("id");
ALTER TABLE "open_board_multi_auth_challenge" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_file_upload" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_board" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_user" ADD FOREIGN KEY ("thumbnail") REFERENCES "open_board_file_upload" ("id");
ALTER TABLE "open_board_general_settings" ADD FOREIGN KEY ("logo") REFERENCES "open_board_file_upload" ("id");
ALTER TABLE "open_board_user" ADD FOREIGN KEY ("external_provider_id") REFERENCES "open_board_external_auth_provider" ("id");
ALTER TABLE "open_board_board_list" ADD FOREIGN KEY ("board_id") REFERENCES "open_board_board" ("id");
ALTER TABLE "open_board_board_label" ADD FOREIGN KEY ("board_id") REFERENCES "open_board_board" ("id");
ALTER TABLE "open_board_board_list_card" ADD FOREIGN KEY ("list_id") REFERENCES "open_board_board_list" ("id");
ALTER TABLE "open_board_board_list_card_comment" ADD FOREIGN KEY ("card_id") REFERENCES "open_board_board_list_card" ("id");
ALTER TABLE "open_board_board_list_card_comment" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
ALTER TABLE "open_board_board_list_card_attachments" ADD FOREIGN KEY ("card_id") REFERENCES "open_board_board_list_card" ("id");
ALTER TABLE "open_board_board_list_card_attachments" ADD FOREIGN KEY ("file_id") REFERENCES "open_board_file_upload" ("id");
ALTER TABLE "open_board_board_list_card_checklist_item" ADD FOREIGN KEY ("card_id") REFERENCES "open_board_board_list_card" ("id");
ALTER TABLE "open_board_board_list_card_card_activity" ADD FOREIGN KEY ("card_id") REFERENCES "open_board_board_list_card" ("id");
ALTER TABLE "open_board_board_list_card_card_activity" ADD FOREIGN KEY ("user_id") REFERENCES "open_board_user" ("id");
CREATE UNIQUE INDEX singleton_auth ON "open_board_auth_settings" ((true));
CREATE UNIQUE INDEX singleton_general ON "open_board_general_settings" ((true));
CREATE UNIQUE INDEX singleton_notifications ON "open_board_notification_settings" ((true));