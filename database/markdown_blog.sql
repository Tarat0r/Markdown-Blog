BEGIN;

CREATE TABLE IF NOT EXISTS public.images (
    id serial NOT NULL,
    hash text COLLATE pg_catalog."default" NOT NULL,
    uploaded_at timestamp without time zone DEFAULT now(),
    CONSTRAINT images_pkey PRIMARY KEY (id),
    CONSTRAINT images_file_hash_key UNIQUE (hash)
);

CREATE TABLE IF NOT EXISTS public.notes (
    id serial NOT NULL,
    user_id integer NOT NULL,
    path character varying(255) COLLATE pg_catalog."default" NOT NULL,
    content text COLLATE pg_catalog."default" NOT NULL,
    hash text COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT notes_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.users (
    id serial NOT NULL,
    name character varying(255) COLLATE pg_catalog."default" NOT NULL,
    email character varying(255) COLLATE pg_catalog."default" NOT NULL,
    api_token text COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_api_token_key UNIQUE (api_token),
    CONSTRAINT users_email_key UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS public.notes_images (
    note_id integer NOT NULL,
    image_id integer NOT NULL,
    CONSTRAINT notes_images_pkey PRIMARY KEY (note_id, image_id),
    CONSTRAINT notes_images_note_id_fkey FOREIGN KEY (note_id)
        REFERENCES public.notes (id) ON DELETE CASCADE,
    CONSTRAINT notes_images_image_id_fkey FOREIGN KEY (image_id)
        REFERENCES public.images (id) ON DELETE CASCADE
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'notes_user_id_fkey'
    ) THEN
        ALTER TABLE public.notes
        ADD CONSTRAINT notes_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE;
    END IF;
END;
$$;


-- Indexes
CREATE INDEX IF NOT EXISTS idx_notes_user_id ON public.notes(user_id);
CREATE INDEX IF NOT EXISTS idx_images_hash ON public.images(hash);
CREATE INDEX IF NOT EXISTS idx_notes_hash ON public.notes(hash);
CREATE INDEX IF NOT EXISTS idx_users_api_token ON public.users(api_token);
CREATE INDEX IF NOT EXISTS idx_notes_images_composite ON public.notes_images(note_id, image_id);

-- Trigger function
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger on notes table
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'trigger_update_modified'
    ) THEN
        CREATE TRIGGER trigger_update_modified
        BEFORE UPDATE ON notes
        FOR EACH ROW
        EXECUTE FUNCTION update_modified_column();
    END IF;
END;
$$;

END;
