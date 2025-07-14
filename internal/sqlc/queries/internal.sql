
-- name: InsertNewLake :one
INSERT INTO lakes (user_id, name, region, ptype)
VALUES ($1, $2, $3, $4)
RETURNING lake_id;

-- name: InsertNewLocation :exec
INSERT INTO locations (lake_id, bucket_name, user_id)
VALUES ($1, $2, $3);


-- name: GetLakeData :one
SELECT 
    lakes.user_id,
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

-- name: DeleteCreds :exec
DELETE 
FROM credentials
WHERE credentials.lake_id = $1;



-- >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>