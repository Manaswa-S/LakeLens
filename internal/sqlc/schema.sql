
CREATE TABLE IF NOT EXISTS public.users
(
    user_id bigint NOT NULL DEFAULT nextval('users_user_id_seq'::regclass),
    email character varying(320) COLLATE pg_catalog."default" NOT NULL,
    confirmed boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_uuid uuid NOT NULL DEFAULT gen_random_uuid(),
    auth_type auth_types NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (user_id),
    CONSTRAINT unique_email UNIQUE (email)
        INCLUDE(email),
    CONSTRAINT unique_user_id UNIQUE (user_id)
        INCLUDE(user_id),
    CONSTRAINT unique_user_uuid UNIQUE (user_uuid)
        INCLUDE(user_uuid)
);

CREATE TABLE IF NOT EXISTS public.goauth
(
    auth_id bigint NOT NULL DEFAULT nextval('goauth_auth_id_seq'::regclass),
    user_id bigint NOT NULL,
    email text COLLATE pg_catalog."default" NOT NULL,
    name text COLLATE pg_catalog."default",
    picture text COLLATE pg_catalog."default",
    id text COLLATE pg_catalog."default",
    CONSTRAINT goauth_pkey PRIMARY KEY (auth_id),
    CONSTRAINT users_goauth_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES public.users (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.epauth
(
    auth_id bigint NOT NULL DEFAULT nextval('epauth_auth_id_seq'::regclass),
    user_id bigint NOT NULL,
    email text COLLATE pg_catalog."default" NOT NULL,
    password text COLLATE pg_catalog."default" NOT NULL,
    name text COLLATE pg_catalog."default",
    picture text COLLATE pg_catalog."default",
    CONSTRAINT epauth_pkey PRIMARY KEY (auth_id),
    CONSTRAINT users_epauth_user_id FOREIGN KEY (user_id)
        REFERENCES public.users (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);


CREATE TABLE IF NOT EXISTS public.settings
(
    set_id bigint NOT NULL DEFAULT nextval('settings_set_id_seq'::regclass),
    user_id bigint NOT NULL,
    last_updated timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    advmeta boolean NOT NULL DEFAULT false,
    cmptview boolean NOT NULL DEFAULT true,
    rfrshint smallint NOT NULL DEFAULT 30,
    notif boolean NOT NULL DEFAULT true,
    theme character varying(25) COLLATE pg_catalog."default" NOT NULL DEFAULT 'light'::character varying,
    fontsz smallint NOT NULL DEFAULT 16,
    tooltps boolean NOT NULL DEFAULT true,
    shortcuts boolean NOT NULL DEFAULT true,
    CONSTRAINT settings_pkey PRIMARY KEY (set_id),
    CONSTRAINT settings_users_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES public.users (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);

CREATE TABLE IF NOT EXISTS public.scans
(
    scan_id bigint NOT NULL DEFAULT nextval('scans_scan_id_seq'::regclass),
    lake_id bigint NOT NULL,
    loc_id bigint NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT scans_pkey PRIMARY KEY (scan_id),
    CONSTRAINT lakes_scans_lake_id FOREIGN KEY (lake_id)
        REFERENCES public.lakes (lake_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    CONSTRAINT locs_scans_loc_id FOREIGN KEY (loc_id)
        REFERENCES public.locations (loc_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
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



CREATE TABLE IF NOT EXISTS public.tips
(
    tip_id bigint NOT NULL DEFAULT nextval('tips_tip_id_seq'::regclass),
    tip text COLLATE pg_catalog."default" NOT NULL,
    hrefs json[],
    CONSTRAINT tips_pkey PRIMARY KEY (tip_id)
);

CREATE TABLE IF NOT EXISTS public.recents
(
    rec_id bigint NOT NULL DEFAULT nextval('recents_rec_id_seq'::regclass),
    user_id bigint NOT NULL,
    action_id bigint NOT NULL,
    "time" timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    action jsonb NOT NULL,
    title text COLLATE pg_catalog."default" NOT NULL,
    description text COLLATE pg_catalog."default",
    CONSTRAINT recents_pkey PRIMARY KEY (rec_id),
    CONSTRAINT users_recents_user_id_fkey FOREIGN KEY (user_id)
        REFERENCES public.users (user_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
);