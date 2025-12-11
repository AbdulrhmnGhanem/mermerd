-- Test case for schema names with hyphens (special characters)
create schema "tenant-alfa";

create table "tenant-alfa".users
(
    id   int          not null primary key,
    name varchar(255) not null
);

create table "tenant-alfa".refresh_tokens
(
    id      int not null primary key,
    user_id int,
    foreign key (user_id) references "tenant-alfa".users (id)
);

create table public.tenants
(
    id   int          not null primary key,
    name varchar(255) not null
);

create table "tenant-alfa".tenant_configs
(
    id        int not null primary key,
    tenant_id int,
    foreign key (tenant_id) references public.tenants (id)
);
