drop table if exists users;
create table users
(
    id uuid primary key,
    mail varchar(255) not null unique ,
    password varchar(255) not null,
    salt varchar(255) not null,
    admin boolean not null default false,
    created_at timestamp not null default current_timestamp,
    last_login timestamp not null default current_timestamp
);

drop table if exists pages;
create table pages
(
    id          varchar(255) unique primary key,
    url         varchar(255) not null,
    title       varchar(255) not null,
    description varchar(255) not null,
    private_page boolean not null default true
);

drop table if exists user_page;
create table user_page
(
    user_id uuid not null,
    page_id varchar(250) not null,
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (page_id) references pages(id) on delete cascade,
    primary key (page_id, user_id)
);

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO java;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public to java;
GRANT ALL PRIVILEGES ON ALL PROCEDURES IN SCHEMA public to java;