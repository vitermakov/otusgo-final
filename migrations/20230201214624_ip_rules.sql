-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    CREATE TYPE public.ip_rules_type as ENUM ('allow', 'deny');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
CREATE TABLE public.ip_rules (
    id uuid NOT NULL,
    type public.ip_rules_type NOT NULL,
    ip_net cidr NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.ip_rules;
DROP TYPE IF EXISTS public.ip_rules_type;
-- +goose StatementEnd
