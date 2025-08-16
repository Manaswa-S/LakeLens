
-- name: GetUserFromEmail :one
SELECT 
    users.user_id,
    users.confirmed,
    users.user_uuid
FROM users 
WHERE users.email = $1;

-- name: GetUserData :one
SELECT 
    users.confirmed,
    users.user_uuid
FROM users 
WHERE users.user_id = $1;



-- name: InsertNewUser :one
INSERT INTO users (email, auth_type) 
VALUES ($1, $2  )
RETURNING users.user_id;

-- name: InsertNewGOAuth :exec
INSERT INTO goauth (user_id, email, name, picture, id)
VALUES ($1, $2, $3, $4, $5);

-- name: InsertNewEPAuth :exec
INSERT INTO epauth (user_id, email, password, name, picture)
VALUES ($1, $2, $3, $4, $5);



-- name: UpdatePass :exec
UPDATE epauth
SET password = $2
WHERE user_id = $1;



-- name: GetGoogleID :one
SELECT
    goauth.auth_id,
    goauth.email,
    goauth.id
FROM goauth
WHERE goauth.user_id = $1;

-- name: GetEPAuthPass :one
SELECT
    epauth.password
FROM epauth
WHERE epauth.user_id = $1;

-- name: CheckIfEPAuth :one
SELECT
    epauth.auth_id
FROM epauth
WHERE epauth.email = $1;


-- name: GetEPDetails :one
SELECT
    epauth.auth_id,
    epauth.email,
    epauth.name,
    epauth.picture
FROM epauth
WHERE epauth.user_id = $1;

-- name: GetGODetails :one
SELECT
    goauth.auth_id,
    goauth.email,
    goauth.name,
    goauth.picture,
    goauth.id
FROM goauth
WHERE goauth.user_id = $1;


-- name: InsertNewSettings :exec
INSERT INTO settings (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO NOTHING;

-- name: UpdateSettings :exec
UPDATE settings
SET 
    last_updated = $2,
    advmeta = $3,
    cmptview = $4,
    rfrshint = $5, 
    notif = $6,
    theme = $7,
    fontsz = $8,
    tooltps = $9,
    shortcuts = $10
WHERE user_id = $1;



-- name: AccDetails :one
SELECT 
    users.user_id,
    users.email,
    users.confirmed,
    users.created_at,
    users.user_uuid,
    users.auth_type::TEXT as auth_type
FROM users
WHERE users.user_id = $1;

-- name: GetLakesList :many
SELECT
    lakes.lake_id,
    lakes.name,
    lakes.region,
    lakes.created_at,
    lakes.ptype
FROM lakes
WHERE lakes.user_id = $1;


-- name: GetLocsList :many
SELECT
    locations.loc_id,
    locations.created_at,
    locations.bucket_name,
    locations.lake_id
FROM locations
WHERE locations.user_id = $1;

-- name: GetLocsListForLake :many
SELECT
    locations.loc_id,
    locations.created_at,
    locations.bucket_name
FROM locations
WHERE locations.user_id = $1 
AND locations.lake_id = $2;



-- name: GetSettings :one
SELECT
    settings.set_id,
    settings.last_updated,
    settings.advmeta,
    settings.cmptview,
    settings.rfrshint,
    settings.notif,
    settings.theme,
    settings.fontsz,
    settings.tooltps,
    settings.shortcuts
FROM settings
WHERE settings.user_id = $1;



-- name: GetLakeDataForUserID :one
SELECT 
    lakes.user_id,
    lakes.name,
    lakes.region,
    lakes.ptype,
    lakes.created_at
FROM lakes 
WHERE lakes.lake_id = $2
AND lakes.user_id = $1;


-- name: DeleteLake :exec
DELETE 
FROM lakes
WHERE lakes.lake_id = $2
AND lakes.user_id = $1;

-- name: DeleteLoc :exec
DELETE
FROM locations
WHERE locations.loc_id = $2
AND locations.user_id = $1;




-- name: InsertNewScan :one
INSERT INTO scans (lake_id, loc_id)
VALUES ($1, $2)
RETURNING scan_id;



-- name: GetTipForID :one
SELECT 
    tips.tip,
    tips.hrefs::json AS hrefs
FROM tips
WHERE tips.tip_id = $1;