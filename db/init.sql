drop table if exists users cascade;
create table users
(
    id uuid primary key,
    mail varchar(255) not null unique ,
    password varchar(255) not null,
    salt bytea not null,
    admin boolean not null default false,
    created_at timestamptz not null default current_timestamp,
    last_login timestamptz not null default current_timestamp,
    blocked boolean not null default false,
    onboard boolean not null default false
);

drop table if exists pages cascade;
create table pages
(
    id          uuid unique primary key,
    technical_name varchar(255) not null unique,
    url         varchar(255) not null,
    title       varchar(40) not null,
    description varchar(355) not null,
    private_page boolean not null default true
);

drop table if exists user_page cascade;
create table user_page
(
    user_id uuid not null,
    page_id uuid not null,
    foreign key (user_id) references users(id) on delete cascade,
    foreign key (page_id) references pages(id) on delete cascade,
    primary key (page_id, user_id)
);

drop table if exists clients cascade;
create table clients
(
    id uuid not null primary key,
    real_user_id uuid null,
    created_at timestamp not null default current_timestamp,
    foreign key (real_user_id) references users(id) on delete cascade
);

drop table if exists client_devices cascade;
create table client_devices
(
    id uuid not null primary key,
    client_id uuid not null,
    ip_address varchar(20) not null,
    user_agent varchar(250) not null,
    ip_location_check_error varchar(255) null,
    created_at timestamp not null default current_timestamp,
    validated bool not null default false,
    foreign key (client_id) references clients(id) on delete cascade,
    constraint device_is_unique_to_client unique (client_id, ip_address, user_agent)
);

drop table if exists ip_locations cascade;
create table ip_locations
(
    device_id uuid primary key,
    city_id int,
    city_name varchar(255),
    city_plz varchar(127),
    subdivision_id int,
    subdivision_code varchar(31),
    country_id int,
    country_code varchar(31),
    continent_id int,
    continent_code varchar(31),
    accuracy_radius int,
    latitude float8,
    longitude float8,
    time_zone varchar(31),
    system_number int,
    system_organisation varchar(127),
    network varchar(255),
    ip_address varchar(31),
    foreign key (device_id) references client_devices(id) on delete cascade
);

drop table if exists user_totp cascade;
create table user_totp
(
    user_id      uuid primary key,
	secret      varchar(255) not null unique,
	account_name varchar(255) not null unique,
	image bytea not null,
    validated   boolean not null default false,
    created_at timestamptz not null default current_timestamp,
    validated_at timestamptz null,
    foreign key (user_id) references users(id) on delete cascade
);
