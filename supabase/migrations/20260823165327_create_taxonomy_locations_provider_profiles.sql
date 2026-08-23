create extension if not exists postgis;

create table public.supported_locales (
  id text primary key,
  active boolean not null default true,
  sort_order integer not null,
  constraint supported_locales_id_length check (char_length(id) between 2 and 10),
  constraint supported_locales_sort_order_nonnegative check (sort_order >= 0)
);

create table public.service_categories (
  id uuid primary key,
  parent_id uuid references public.service_categories(id) on delete restrict,
  slug text not null unique,
  active boolean not null default true,
  sort_order integer not null,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint service_categories_parent_not_self check (parent_id is null or parent_id <> id),
  constraint service_categories_slug_length check (char_length(slug) between 1 and 80),
  constraint service_categories_sort_order_nonnegative check (sort_order >= 0)
);

create index service_categories_parent_sort_idx
  on public.service_categories (parent_id, sort_order, id);

create table public.service_category_translations (
  category_id uuid not null references public.service_categories(id) on delete cascade,
  locale text not null references public.supported_locales(id) on delete restrict,
  name text not null,
  description text,
  primary key (category_id, locale),
  constraint service_category_translation_name_length check (char_length(name) between 1 and 120),
  constraint service_category_translation_description_length check (description is null or char_length(description) <= 500)
);

create table public.spoken_languages (
  id text primary key,
  active boolean not null default true,
  sort_order integer not null,
  constraint spoken_languages_id_length check (char_length(id) between 2 and 10),
  constraint spoken_languages_sort_order_nonnegative check (sort_order >= 0)
);

create table public.spoken_language_translations (
  language_code text not null references public.spoken_languages(id) on delete cascade,
  locale text not null references public.supported_locales(id) on delete restrict,
  name text not null,
  primary key (language_code, locale),
  constraint spoken_language_translation_name_length check (char_length(name) between 1 and 80)
);

create table public.administrative_areas (
  id uuid primary key,
  source text not null,
  source_version text not null,
  external_code text not null,
  kind text not null,
  name text not null,
  parent_id uuid references public.administrative_areas(id) on delete restrict,
  active boolean not null default true,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint administrative_areas_source_length check (char_length(source) between 1 and 40),
  constraint administrative_areas_source_version_length check (char_length(source_version) between 1 and 20),
  constraint administrative_areas_external_code_length check (char_length(external_code) between 1 and 32),
  constraint administrative_areas_kind check (kind in ('country', 'district', 'municipality', 'parish')),
  constraint administrative_areas_name_length check (char_length(name) between 1 and 160),
  constraint administrative_areas_parent_not_self check (parent_id is null or parent_id <> id),
  unique (source, external_code)
);

create index administrative_areas_parent_kind_idx
  on public.administrative_areas (parent_id, kind, id);

create table public.localities (
  id uuid primary key,
  slug text not null unique,
  name text not null,
  parent_parish_id uuid not null references public.administrative_areas(id) on delete restrict,
  source text not null,
  source_element_id text not null unique,
  source_version text not null,
  source_retrieved_at timestamptz not null,
  latitude double precision not null,
  longitude double precision not null,
  center geography(point, 4326) generated always as
    (st_setsrid(st_makepoint(longitude, latitude), 4326)::geography) stored,
  active boolean not null default true,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint localities_slug_length check (char_length(slug) between 1 and 100),
  constraint localities_name_length check (char_length(name) between 1 and 160),
  constraint localities_source_length check (char_length(source) between 1 and 40),
  constraint localities_source_element_id_length check (char_length(source_element_id) between 2 and 32),
  constraint localities_source_version_length check (char_length(source_version) between 1 and 20),
  constraint localities_latitude_range check (latitude between -90 and 90),
  constraint localities_longitude_range check (longitude between -180 and 180)
);

create index localities_parent_name_idx
  on public.localities (parent_parish_id, name, id);
create index localities_center_gist_idx
  on public.localities using gist (center);

create table public.provider_profiles (
  internal_user_id uuid primary key references public.user_accounts(internal_user_id) on delete cascade,
  display_name text not null,
  provider_type text not null,
  bio text not null,
  primary_locality_id uuid not null references public.localities(id) on delete restrict,
  max_travel_distance_km integer not null,
  travels_to_customer boolean not null default false,
  receives_customer boolean not null default false,
  remote_services boolean not null default false,
  created_at timestamptz not null default timezone('utc', now()),
  updated_at timestamptz not null default timezone('utc', now()),
  constraint provider_profiles_display_name_length check (char_length(btrim(display_name)) between 2 and 100),
  constraint provider_profiles_type check (provider_type in ('individual', 'professional', 'business')),
  constraint provider_profiles_bio_length check (char_length(bio) <= 1000),
  constraint provider_profiles_travel_distance check (max_travel_distance_km between 0 and 200),
  constraint provider_profiles_service_mode check (travels_to_customer or receives_customer or remote_services),
  constraint provider_profiles_zero_travel_mode check (max_travel_distance_km > 0 or receives_customer or remote_services)
);

create index provider_profiles_primary_locality_idx
  on public.provider_profiles (primary_locality_id, internal_user_id);

create table public.provider_service_localities (
  internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete cascade,
  locality_id uuid not null references public.localities(id) on delete restrict,
  primary key (internal_user_id, locality_id)
);

create index provider_service_localities_locality_idx
  on public.provider_service_localities (locality_id, internal_user_id);

create table public.provider_spoken_languages (
  internal_user_id uuid not null references public.provider_profiles(internal_user_id) on delete cascade,
  language_code text not null references public.spoken_languages(id) on delete restrict,
  primary key (internal_user_id, language_code)
);

create index provider_spoken_languages_language_idx
  on public.provider_spoken_languages (language_code, internal_user_id);

insert into public.supported_locales (id, active, sort_order) values
  ('pt-PT', true, 10),
  ('en', true, 20),
  ('es', true, 30);

insert into public.service_categories (id, parent_id, slug, active, sort_order) values
  (md5('juntly:category:home-repairs')::uuid, null, 'home-repairs', true, 10),
  (md5('juntly:category:plumbing')::uuid, md5('juntly:category:home-repairs')::uuid, 'plumbing', true, 10),
  (md5('juntly:category:electrical-work')::uuid, md5('juntly:category:home-repairs')::uuid, 'electrical-work', true, 20),
  (md5('juntly:category:construction')::uuid, md5('juntly:category:home-repairs')::uuid, 'construction', true, 30),
  (md5('juntly:category:small-repairs')::uuid, md5('juntly:category:home-repairs')::uuid, 'small-repairs', true, 40),
  (md5('juntly:category:home-and-garden')::uuid, null, 'home-and-garden', true, 20),
  (md5('juntly:category:cleaning')::uuid, md5('juntly:category:home-and-garden')::uuid, 'cleaning', true, 10),
  (md5('juntly:category:gardening')::uuid, md5('juntly:category:home-and-garden')::uuid, 'gardening', true, 20),
  (md5('juntly:category:rural-and-transport')::uuid, null, 'rural-and-transport', true, 30),
  (md5('juntly:category:agricultural-assistance')::uuid, md5('juntly:category:rural-and-transport')::uuid, 'agricultural-assistance', true, 10),
  (md5('juntly:category:transport')::uuid, md5('juntly:category:rural-and-transport')::uuid, 'transport', true, 20),
  (md5('juntly:category:care-and-learning')::uuid, null, 'care-and-learning', true, 40),
  (md5('juntly:category:elderly-assistance')::uuid, md5('juntly:category:care-and-learning')::uuid, 'elderly-assistance', true, 10),
  (md5('juntly:category:animal-care')::uuid, md5('juntly:category:care-and-learning')::uuid, 'animal-care', true, 20),
  (md5('juntly:category:private-lessons')::uuid, md5('juntly:category:care-and-learning')::uuid, 'private-lessons', true, 30),
  (md5('juntly:category:food-and-technology')::uuid, null, 'food-and-technology', true, 50),
  (md5('juntly:category:meal-preparation')::uuid, md5('juntly:category:food-and-technology')::uuid, 'meal-preparation', true, 10),
  (md5('juntly:category:computer-repair')::uuid, md5('juntly:category:food-and-technology')::uuid, 'computer-repair', true, 20);

with category_names(slug, locale, name) as (values
  ('home-repairs', 'pt-PT', 'Reparações domésticas'), ('home-repairs', 'en', 'Home repairs'), ('home-repairs', 'es', 'Reparaciones del hogar'),
  ('plumbing', 'pt-PT', 'Canalização'), ('plumbing', 'en', 'Plumbing'), ('plumbing', 'es', 'Fontanería'),
  ('electrical-work', 'pt-PT', 'Eletricidade'), ('electrical-work', 'en', 'Electrical work'), ('electrical-work', 'es', 'Electricidad'),
  ('construction', 'pt-PT', 'Construção'), ('construction', 'en', 'Construction'), ('construction', 'es', 'Construcción'),
  ('small-repairs', 'pt-PT', 'Pequenas reparações'), ('small-repairs', 'en', 'Small repairs'), ('small-repairs', 'es', 'Pequeñas reparaciones'),
  ('home-and-garden', 'pt-PT', 'Casa e jardim'), ('home-and-garden', 'en', 'Home and garden'), ('home-and-garden', 'es', 'Hogar y jardín'),
  ('cleaning', 'pt-PT', 'Limpeza'), ('cleaning', 'en', 'Cleaning'), ('cleaning', 'es', 'Limpieza'),
  ('gardening', 'pt-PT', 'Jardinagem'), ('gardening', 'en', 'Gardening'), ('gardening', 'es', 'Jardinería'),
  ('rural-and-transport', 'pt-PT', 'Serviços rurais e transporte'), ('rural-and-transport', 'en', 'Rural services and transport'), ('rural-and-transport', 'es', 'Servicios rurales y transporte'),
  ('agricultural-assistance', 'pt-PT', 'Apoio agrícola'), ('agricultural-assistance', 'en', 'Agricultural assistance'), ('agricultural-assistance', 'es', 'Ayuda agrícola'),
  ('transport', 'pt-PT', 'Transporte'), ('transport', 'en', 'Transport'), ('transport', 'es', 'Transporte'),
  ('care-and-learning', 'pt-PT', 'Cuidados e aprendizagem'), ('care-and-learning', 'en', 'Care and learning'), ('care-and-learning', 'es', 'Cuidados y aprendizaje'),
  ('elderly-assistance', 'pt-PT', 'Apoio a idosos'), ('elderly-assistance', 'en', 'Elderly assistance'), ('elderly-assistance', 'es', 'Ayuda a mayores'),
  ('animal-care', 'pt-PT', 'Cuidados de animais'), ('animal-care', 'en', 'Animal care'), ('animal-care', 'es', 'Cuidado de animales'),
  ('private-lessons', 'pt-PT', 'Aulas particulares'), ('private-lessons', 'en', 'Private lessons'), ('private-lessons', 'es', 'Clases particulares'),
  ('food-and-technology', 'pt-PT', 'Alimentação e tecnologia'), ('food-and-technology', 'en', 'Food and technology'), ('food-and-technology', 'es', 'Alimentación y tecnología'),
  ('meal-preparation', 'pt-PT', 'Preparação de refeições'), ('meal-preparation', 'en', 'Meal preparation'), ('meal-preparation', 'es', 'Preparación de comidas'),
  ('computer-repair', 'pt-PT', 'Reparação de computadores'), ('computer-repair', 'en', 'Computer repair'), ('computer-repair', 'es', 'Reparación de ordenadores')
)
insert into public.service_category_translations (category_id, locale, name)
select categories.id, category_names.locale, category_names.name
from category_names
join public.service_categories categories on categories.slug = category_names.slug;

insert into public.spoken_languages (id, active, sort_order) values
  ('pt-PT', true, 10),
  ('en', true, 20),
  ('es', true, 30);

insert into public.spoken_language_translations (language_code, locale, name) values
  ('pt-PT', 'pt-PT', 'Português'), ('pt-PT', 'en', 'Portuguese'), ('pt-PT', 'es', 'Portugués'),
  ('en', 'pt-PT', 'Inglês'), ('en', 'en', 'English'), ('en', 'es', 'Inglés'),
  ('es', 'pt-PT', 'Espanhol'), ('es', 'en', 'Spanish'), ('es', 'es', 'Español');

insert into public.administrative_areas
  (id, source, source_version, external_code, kind, name, parent_id, active) values
  (md5('juntly:area:caop:PT')::uuid, 'caop', '2025', 'PT', 'country', 'Portugal', null, true),
  (md5('juntly:area:caop:05')::uuid, 'caop', '2025', '05', 'district', 'Castelo Branco', md5('juntly:area:caop:PT')::uuid, true),
  (md5('juntly:area:caop:0502')::uuid, 'caop', '2025', '0502', 'municipality', 'Castelo Branco', md5('juntly:area:caop:05')::uuid, true),
  (md5('juntly:area:caop:0505')::uuid, 'caop', '2025', '0505', 'municipality', 'Idanha-a-Nova', md5('juntly:area:caop:05')::uuid, true),
  (md5('juntly:area:caop:050205')::uuid, 'caop', '2025', '050205', 'parish', 'Castelo Branco', md5('juntly:area:caop:0502')::uuid, true),
  (md5('juntly:area:caop:050510')::uuid, 'caop', '2025', '050510', 'parish', 'Penha Garcia', md5('juntly:area:caop:0505')::uuid, true),
  (md5('juntly:area:caop:050518')::uuid, 'caop', '2025', '050518', 'parish', 'União das freguesias de Idanha-a-Nova e Alcafozes', md5('juntly:area:caop:0505')::uuid, true),
  (md5('juntly:area:caop:050520')::uuid, 'caop', '2025', '050520', 'parish', 'União das freguesias de Monsanto e Idanha-a-Velha', md5('juntly:area:caop:0505')::uuid, true),
  (md5('juntly:area:caop:050521')::uuid, 'caop', '2025', '050521', 'parish', 'União das freguesias de Zebreira e Segura', md5('juntly:area:caop:0505')::uuid, true);

insert into public.localities
  (id, slug, name, parent_parish_id, source, source_element_id, source_version, source_retrieved_at, latitude, longitude, active) values
  (md5('juntly:locality:castelo-branco')::uuid, 'castelo-branco', 'Castelo Branco', md5('juntly:area:caop:050205')::uuid, 'OpenStreetMap', 'R5396187', '2026-08-23', '2026-08-23T00:00:00Z', 39.8266322, -7.4919318, true),
  (md5('juntly:locality:idanha-a-nova')::uuid, 'idanha-a-nova', 'Idanha-a-Nova', md5('juntly:area:caop:050518')::uuid, 'OpenStreetMap', 'R5395738', '2026-08-23', '2026-08-23T00:00:00Z', 39.9260883, -7.2436356, true),
  (md5('juntly:locality:monsanto')::uuid, 'monsanto', 'Monsanto', md5('juntly:area:caop:050520')::uuid, 'OpenStreetMap', 'N371426674', '2026-08-23', '2026-08-23T00:00:00Z', 40.0387510, -7.1151133, true),
  (md5('juntly:locality:penha-garcia')::uuid, 'penha-garcia', 'Penha Garcia', md5('juntly:area:caop:050510')::uuid, 'OpenStreetMap', 'R5431477', '2026-08-23', '2026-08-23T00:00:00Z', 40.0422569, -7.0163521, true),
  (md5('juntly:locality:zebreira')::uuid, 'zebreira', 'Zebreira', md5('juntly:area:caop:050521')::uuid, 'OpenStreetMap', 'N440173641', '2026-08-23', '2026-08-23T00:00:00Z', 39.8455920, -7.0703366, true);
