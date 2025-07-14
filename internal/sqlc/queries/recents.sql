
-- name: InsertRecentsBulk :exec
INSERT INTO recents (user_id, action_id, time, action, title, description)
SELECT UNNEST($1::bigint[]), UNNEST($2::bigint[]), UNNEST($3::timestamptz[]), UNNEST($4::jsonb[]), UNNEST($5::text[]), UNNEST($6::text[]);


-- name: GetRecents :many
SELECT 
    recents.rec_id,
    recents.action_id,
    recents.time,
    recents.action
FROM recents
WHERE recents.user_id = $1
LIMIT $2
OFFSET $3; 

-- name: ResolveLocID :one
SELECT
    locations.bucket_name
FROM locations
WHERE locations.loc_id = $1;


-- name: UnResolveLakeName :one
SELECT
    lakes.lake_id
FROM lakes
WHERE lakes.user_id = $1
AND lakes.name = $2;

-- name: ResolveDescNewLake :one
SELECT
    COUNT(locations.loc_id)
FROM locations
WHERE locations.lake_id = $1;