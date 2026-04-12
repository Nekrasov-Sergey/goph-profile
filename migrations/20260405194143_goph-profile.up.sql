do
$$
    begin
        create type processing_status as enum ('pending', 'processing', 'completed', 'failed');
    exception
        when duplicate_object then null;
    end
$$;

create table if not exists avatars
(
    id                uuid primary key  default gen_random_uuid(),
    user_id           varchar(255) not null,
    file_name         varchar(255) not null,
    mime_type         varchar(100) not null,
    size_bytes        bigint       not null,
    width             integer,
    height            integer,
    s3_key            varchar(500) not null,
    thumbnail_s3_keys jsonb,
    processing_status processing_status default 'pending',
    created_at        timestamptz       default now(),
    updated_at        timestamptz       default now(),
    deleted_at        timestamptz
);

create index if not exists avatars_user_id_idx on avatars (user_id) where deleted_at is null;
create index if not exists avatars_processing_status_idx on avatars (processing_status);
