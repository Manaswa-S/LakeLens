
-- name: InsertNewCredentails :exec
INSERT INTO credentials (key_id, key, region)
VALUES ($1, $2, $3);


-- name: GetCredentials :one
SELECT * 
FROM credentials 
WHERE cred_id = $1; 