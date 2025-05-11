
CREATE TABLE IF NOT EXISTS public.users
(
    user_id bigint NOT NULL DEFAULT nextval('users_user_id_seq'::regclass),
    email character varying(320) COLLATE pg_catalog."default" NOT NULL,
    password character varying(72) COLLATE pg_catalog."default" NOT NULL,
    confirmed boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_uuid uuid NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT users_pkey PRIMARY KEY (user_id),
    CONSTRAINT unique_email UNIQUE (email)
        INCLUDE(email),
    CONSTRAINT unique_user_id UNIQUE (user_id)
        INCLUDE(user_id),
    CONSTRAINT unique_user_uuid UNIQUE (user_uuid)
        INCLUDE(user_uuid)
);

CREATE TABLE IF NOT EXISTS public.lakes
(
    lake_id bigint NOT NULL DEFAULT nextval('lakes_lake_id_seq'::regclass),
    user_id bigint NOT NULL,
    name text COLLATE pg_catalog."default" NOT NULL,
    region text COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ptype text COLLATE pg_catalog."default" NOT NULL DEFAULT ''::text,
    CONSTRAINT lakes_pkey PRIMARY KEY (lake_id),
    CONSTRAINT users_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES public.users (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.credentials
(
    cred_id bigint NOT NULL DEFAULT nextval('credentials_cred_id_seq'::regclass),
    lake_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    key_id text COLLATE pg_catalog."default" NOT NULL,
    key text COLLATE pg_catalog."default" NOT NULL,
    region text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT credentials_pkey PRIMARY KEY (cred_id),
    CONSTRAINT unique_key_id UNIQUE (key_id)
        INCLUDE(key_id),
    CONSTRAINT lakes_lake_id_fkey FOREIGN KEY (lake_id)
        REFERENCES public.lakes (lake_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);


CREATE TABLE IF NOT EXISTS public.locations
(
    loc_id bigint NOT NULL DEFAULT nextval('locations_loc_id_seq'::regclass),
    lake_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bucket_name text COLLATE pg_catalog."default" NOT NULL,
    user_id bigint NOT NULL DEFAULT 1,
    CONSTRAINT locations_pkey PRIMARY KEY (loc_id),
    CONSTRAINT lakes_lake_id_fkey FOREIGN KEY (lake_id)
        REFERENCES public.lakes (lake_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT users_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES public.users (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);