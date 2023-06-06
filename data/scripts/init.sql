drop table if exists users;
create table users
(
    id serial primary key,
    mail varchar(255) not null unique ,
    password varchar(255) not null,
    salt varchar(255) not null,
    admin boolean not null default false,
    view_private boolean not null default false,
    created_at timestamp not null default current_timestamp,
    last_login timestamp not null default current_timestamp
);

drop table if exists pages;
create table pages
(
    id          serial primary key,
    url         varchar(255) not null,
    title       varchar(255) not null,
    description varchar(255) not null,
    private_access boolean not null default true
);

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO java;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public to java;
GRANT ALL PRIVILEGES ON ALL PROCEDURES IN SCHEMA public to java;