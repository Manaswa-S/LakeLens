
-- name: InsertNewUser :exec
INSERT INTO users (email, password) 
VALUES ($1, $2);





-- name: InsertNewLake :one
INSERT INTO lakes (user_id, name, region)
VALUES ($1, $2, $3)
RETURNING lake_id;



-- name: GetLakeIDfromLocID :one
SELECT 
    locations.lake_id
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