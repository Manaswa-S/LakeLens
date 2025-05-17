
-- name: GetUserFromEmail :one
SELECT 
    users.user_id,
    users.password,
    users.confirmed,
    users.user_uuid
FROM users 
WHERE users.email = $1;


-- name: GetPassForUserID :one
SELECT 
    users.password
FROM users 
WHERE users.user_id = $1;

-- name: GetUserData :one
SELECT 
    users.confirmed,
    users.user_uuid
FROM users 
WHERE users.user_id = $1;


-- name: CheckIfEPassOnly :one
SELECT
    users.user_id
FROM users
WHERE NOT EXISTS (
    SELECT 1 FROM goauth WHERE users.user_id = goauth.user_id
)
AND users.email = $1;


-- name: UpdatePass :exec
UPDATE users
SET password = $2
WHERE user_id = $1;



-- name: InsertNewUser :one
INSERT INTO users (email, password) 
VALUES ($1, $2)
RETURNING users.user_id;



-- name: InsertNewGOAuth :exec
INSERT INTO goauth (user_id, email, name, picture, id)
VALUES ($1, $2, $3, $4, $5);


-- name: InsertNewLake :one
INSERT INTO lakes (user_id, name, region, ptype)
VALUES ($1, $2, $3, $4)
RETURNING lake_id;



-- name: GetGoogleID :one
SELECT
    goauth.auth_id,
    goauth.email,
    goauth.id
FROM goauth
WHERE goauth.user_id = $1;



-- name: GetLakeData :one
SELECT 
    lakes.name,
    lakes.region,
    lakes.ptype
FROM lakes 
WHERE lakes.lake_id = $1;


-- name: GetLocationData :one
SELECT 
    locations.loc_id,
    locations.lake_id,
    locations.bucket_name,
    locations.user_id
FROM locations 
WHERE loc_id = $1;



-- name: InsertNewCredentails :exec
INSERT INTO credentials (lake_id, key_id, key, region)
VALUES ($1, $2, $3, $4);


-- name: GetCredentials :one
SELECT 
    credentials.key_id,
    credentials.key,
    credentials.region
FROM credentials 
WHERE lake_id = $1; 