-- +migrate Up
ALTER TABLE user_preferences DROP COLUMN theme_toggle_enabled;

INSERT INTO cvars (
  cvar_key, cvar_value_int, cvar_value_string, cvar_main_type,
  cvar_title, cvar_description, cvar_category, cvar_ordinal
)
SELECT
  'theme_name', NULL, cvar_value_string, 'string',
  'Theme name',
  'Default or enforced drop-in CSS theme (empty = Femtocrank base styling only).',
  'Theme', 20
FROM cvars
WHERE cvar_key = 'custom_theme'
  AND NOT EXISTS (SELECT 1 FROM cvars WHERE cvar_key = 'theme_name');

DELETE FROM cvars WHERE cvar_key IN ('theme_mode', 'custom_theme');

-- +migrate Down
ALTER TABLE user_preferences
  ADD COLUMN theme_toggle_enabled TINYINT(1) NOT NULL DEFAULT 0 AFTER sidebar_enabled;

DELETE FROM cvars WHERE cvar_key IN (
  'theme_color_scheme_switcher_enabled',
  'theme_name',
  'theme_control'
);

INSERT INTO cvars (
  cvar_key, cvar_value_int, cvar_value_string, cvar_main_type,
  cvar_title, cvar_description, cvar_category, cvar_ordinal
) VALUES
  ('theme_mode', NULL, 'auto', 'string', 'Color scheme',
   'Light/dark appearance for all users. Auto follows the browser or system preference.',
   'Site', 25),
  ('custom_theme', NULL, '', 'string', 'Drop-in theme',
   'Optional PicoCrank supplemental theme layered on Femtocrank for all users. Default uses Femtocrank only.',
   'Site', 26);
