
CREATE TABLE public.credentials
(
    cred_id BIGINT NOT NULL DEFAULT nextval('credentials_cred_id_seq'::regclass),
    key_id TEXT NOT NULL,
    key TEXT NOT NULL,
    region TEXT NOT NULL,
    CONSTRAINT credentials_pkey PRIMARY KEY (cred_id)
);