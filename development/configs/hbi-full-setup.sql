--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4
-- Dumped by pg_dump version 17.5

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


SELECT 'CREATE DATABASE "host-inventory"' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'host-inventory')\gexec
-- CREATE DATABASE IF NOT EXISTS "host-inventory";
CREATE ROLE "VBzEIpnveJmm3TXi" with LOGIN PASSWORD 'mypassword';
GRANT ALL PRIVILEGES on database "host-inventory" to "VBzEIpnveJmm3TXi";
ALTER ROLE "VBzEIpnveJmm3TXi" with superuser;

--
-- Name: hbi; Type: SCHEMA; Schema: -; Owner: VBzEIpnveJmm3TXi
--

CREATE SCHEMA hbi;


ALTER SCHEMA hbi OWNER TO "VBzEIpnveJmm3TXi";

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: alembic_version; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.alembic_version (
    version_num character varying(32) NOT NULL
);


ALTER TABLE hbi.alembic_version OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: groups; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.groups (
    id uuid NOT NULL,
    org_id character varying(36) NOT NULL,
    account character varying(10),
    name character varying(255) NOT NULL,
    created_on timestamp with time zone NOT NULL,
    modified_on timestamp with time zone NOT NULL,
    ungrouped boolean DEFAULT false NOT NULL
);


ALTER TABLE hbi.groups OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hbi_metadata; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hbi_metadata (
    name character varying NOT NULL,
    type character varying NOT NULL,
    last_succeeded timestamp with time zone NOT NULL
);


ALTER TABLE hbi.hbi_metadata OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts (
    id uuid NOT NULL,
    account character varying(10),
    display_name character varying(200),
    created_on timestamp with time zone NOT NULL,
    modified_on timestamp with time zone NOT NULL,
    facts jsonb,
    tags jsonb,
    canonical_facts jsonb NOT NULL,
    system_profile_facts jsonb,
    ansible_host character varying(255),
    stale_timestamp timestamp with time zone NOT NULL,
    reporter character varying(255) NOT NULL,
    per_reporter_staleness jsonb DEFAULT '{}'::jsonb NOT NULL,
    org_id character varying(36) NOT NULL,
    groups jsonb NOT NULL,
    tags_alt jsonb,
    last_check_in timestamp with time zone,
    stale_warning_timestamp timestamp with time zone,
    deletion_timestamp timestamp with time zone
);


ALTER TABLE hbi.hosts OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_groups; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts_groups (
    group_id uuid NOT NULL,
    host_id uuid NOT NULL
);


ALTER TABLE hbi.hosts_groups OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_groups_new; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts_groups_new (
    org_id character varying(36) NOT NULL,
    host_id uuid NOT NULL,
    group_id uuid NOT NULL
)
PARTITION BY HASH (org_id);


ALTER TABLE hbi.hosts_groups_new OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_groups_p0; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts_groups_p0 (
    org_id character varying(36) NOT NULL,
    host_id uuid NOT NULL,
    group_id uuid NOT NULL
);


ALTER TABLE hbi.hosts_groups_p0 OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_groups_p1; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts_groups_p1 (
    org_id character varying(36) NOT NULL,
    host_id uuid NOT NULL,
    group_id uuid NOT NULL
);


ALTER TABLE hbi.hosts_groups_p1 OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_new; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts_new (
    org_id character varying(36) NOT NULL,
    id uuid NOT NULL,
    account character varying(10),
    display_name character varying(200),
    ansible_host character varying(255),
    created_on timestamp with time zone,
    modified_on timestamp with time zone,
    facts jsonb,
    tags jsonb,
    tags_alt jsonb,
    system_profile_facts jsonb,
    groups jsonb NOT NULL,
    last_check_in timestamp with time zone,
    stale_timestamp timestamp with time zone NOT NULL,
    deletion_timestamp timestamp with time zone,
    stale_warning_timestamp timestamp with time zone,
    reporter character varying(255) NOT NULL,
    per_reporter_staleness jsonb NOT NULL,
    canonical_facts jsonb NOT NULL,
    canonical_facts_version integer,
    is_virtual boolean,
    insights_id uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid NOT NULL,
    subscription_manager_id character varying(36),
    satellite_id character varying(255),
    fqdn character varying(255),
    bios_uuid character varying(36),
    ip_addresses jsonb,
    mac_addresses jsonb,
    provider_id character varying(500),
    provider_type character varying(50)
)
PARTITION BY HASH (org_id);


ALTER TABLE hbi.hosts_new OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_p0; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts_p0 (
    org_id character varying(36) NOT NULL,
    id uuid NOT NULL,
    account character varying(10),
    display_name character varying(200),
    ansible_host character varying(255),
    created_on timestamp with time zone,
    modified_on timestamp with time zone,
    facts jsonb,
    tags jsonb,
    tags_alt jsonb,
    system_profile_facts jsonb,
    groups jsonb NOT NULL,
    last_check_in timestamp with time zone,
    stale_timestamp timestamp with time zone NOT NULL,
    deletion_timestamp timestamp with time zone,
    stale_warning_timestamp timestamp with time zone,
    reporter character varying(255) NOT NULL,
    per_reporter_staleness jsonb NOT NULL,
    canonical_facts jsonb NOT NULL,
    canonical_facts_version integer,
    is_virtual boolean,
    insights_id uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid NOT NULL,
    subscription_manager_id character varying(36),
    satellite_id character varying(255),
    fqdn character varying(255),
    bios_uuid character varying(36),
    ip_addresses jsonb,
    mac_addresses jsonb,
    provider_id character varying(500),
    provider_type character varying(50)
);


ALTER TABLE hbi.hosts_p0 OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_p1; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.hosts_p1 (
    org_id character varying(36) NOT NULL,
    id uuid NOT NULL,
    account character varying(10),
    display_name character varying(200),
    ansible_host character varying(255),
    created_on timestamp with time zone,
    modified_on timestamp with time zone,
    facts jsonb,
    tags jsonb,
    tags_alt jsonb,
    system_profile_facts jsonb,
    groups jsonb NOT NULL,
    last_check_in timestamp with time zone,
    stale_timestamp timestamp with time zone NOT NULL,
    deletion_timestamp timestamp with time zone,
    stale_warning_timestamp timestamp with time zone,
    reporter character varying(255) NOT NULL,
    per_reporter_staleness jsonb NOT NULL,
    canonical_facts jsonb NOT NULL,
    canonical_facts_version integer,
    is_virtual boolean,
    insights_id uuid DEFAULT '00000000-0000-0000-0000-000000000000'::uuid NOT NULL,
    subscription_manager_id character varying(36),
    satellite_id character varying(255),
    fqdn character varying(255),
    bios_uuid character varying(36),
    ip_addresses jsonb,
    mac_addresses jsonb,
    provider_id character varying(500),
    provider_type character varying(50)
);


ALTER TABLE hbi.hosts_p1 OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: staleness; Type: TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE TABLE hbi.staleness (
    id uuid NOT NULL,
    org_id character varying(36) NOT NULL,
    conventional_time_to_stale integer NOT NULL,
    conventional_time_to_stale_warning integer NOT NULL,
    conventional_time_to_delete integer NOT NULL,
    immutable_time_to_stale integer NOT NULL,
    immutable_time_to_stale_warning integer NOT NULL,
    immutable_time_to_delete integer NOT NULL,
    created_on timestamp with time zone NOT NULL,
    modified_on timestamp with time zone NOT NULL
);


ALTER TABLE hbi.staleness OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hosts_groups_p0; Type: TABLE ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups_new ATTACH PARTITION hbi.hosts_groups_p0 FOR VALUES WITH (modulus 2, remainder 0);


--
-- Name: hosts_groups_p1; Type: TABLE ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups_new ATTACH PARTITION hbi.hosts_groups_p1 FOR VALUES WITH (modulus 2, remainder 1);


--
-- Name: hosts_p0; Type: TABLE ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_new ATTACH PARTITION hbi.hosts_p0 FOR VALUES WITH (modulus 2, remainder 0);


--
-- Name: hosts_p1; Type: TABLE ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_new ATTACH PARTITION hbi.hosts_p1 FOR VALUES WITH (modulus 2, remainder 1);


--
-- Data for Name: alembic_version; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.alembic_version (version_num) FROM stdin;
28280de3f1ce
\.


--
-- Data for Name: groups; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.groups (id, org_id, account, name, created_on, modified_on, ungrouped) FROM stdin;
\.


--
-- Data for Name: hbi_metadata; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.hbi_metadata (name, type, last_succeeded) FROM stdin;
\.


--
-- Data for Name: hosts; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.hosts (id, account, display_name, created_on, modified_on, facts, tags, canonical_facts, system_profile_facts, ansible_host, stale_timestamp, reporter, per_reporter_staleness, org_id, groups, tags_alt, last_check_in, stale_warning_timestamp, deletion_timestamp) FROM stdin;
\.


--
-- Data for Name: hosts_groups; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.hosts_groups (group_id, host_id) FROM stdin;
\.


--
-- Data for Name: hosts_groups_p0; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.hosts_groups_p0 (org_id, host_id, group_id) FROM stdin;
\.


--
-- Data for Name: hosts_groups_p1; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.hosts_groups_p1 (org_id, host_id, group_id) FROM stdin;
\.


--
-- Data for Name: hosts_p0; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.hosts_p0 (org_id, id, account, display_name, ansible_host, created_on, modified_on, facts, tags, tags_alt, system_profile_facts, groups, last_check_in, stale_timestamp, deletion_timestamp, stale_warning_timestamp, reporter, per_reporter_staleness, canonical_facts, canonical_facts_version, is_virtual, insights_id, subscription_manager_id, satellite_id, fqdn, bios_uuid, ip_addresses, mac_addresses, provider_id, provider_type) FROM stdin;
\.


--
-- Data for Name: hosts_p1; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.hosts_p1 (org_id, id, account, display_name, ansible_host, created_on, modified_on, facts, tags, tags_alt, system_profile_facts, groups, last_check_in, stale_timestamp, deletion_timestamp, stale_warning_timestamp, reporter, per_reporter_staleness, canonical_facts, canonical_facts_version, is_virtual, insights_id, subscription_manager_id, satellite_id, fqdn, bios_uuid, ip_addresses, mac_addresses, provider_id, provider_type) FROM stdin;
\.


--
-- Data for Name: staleness; Type: TABLE DATA; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

COPY hbi.staleness (id, org_id, conventional_time_to_stale, conventional_time_to_stale_warning, conventional_time_to_delete, immutable_time_to_stale, immutable_time_to_stale_warning, immutable_time_to_delete, created_on, modified_on) FROM stdin;
\.


--
-- Name: alembic_version alembic_version_pkc; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.alembic_version
    ADD CONSTRAINT alembic_version_pkc PRIMARY KEY (version_num);


--
-- Name: groups groups_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.groups
    ADD CONSTRAINT groups_pkey PRIMARY KEY (id);

ALTER TABLE ONLY hbi.groups REPLICA IDENTITY USING INDEX groups_pkey;


--
-- Name: hbi_metadata hbi_metadata_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hbi_metadata
    ADD CONSTRAINT hbi_metadata_pkey PRIMARY KEY (name, type);


--
-- Name: hosts_groups_new hosts_groups_new_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups_new
    ADD CONSTRAINT hosts_groups_new_pkey PRIMARY KEY (org_id, host_id, group_id);


--
-- Name: hosts_groups_p0 hosts_groups_p0_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups_p0
    ADD CONSTRAINT hosts_groups_p0_pkey PRIMARY KEY (org_id, host_id, group_id);


--
-- Name: hosts_groups_p1 hosts_groups_p1_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups_p1
    ADD CONSTRAINT hosts_groups_p1_pkey PRIMARY KEY (org_id, host_id, group_id);


--
-- Name: hosts_groups hosts_groups_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups
    ADD CONSTRAINT hosts_groups_pkey PRIMARY KEY (group_id, host_id);

ALTER TABLE ONLY hbi.hosts_groups REPLICA IDENTITY USING INDEX hosts_groups_pkey;


--
-- Name: hosts_groups hosts_groups_unique_host_id; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups
    ADD CONSTRAINT hosts_groups_unique_host_id UNIQUE (host_id);


--
-- Name: hosts_new hosts_new_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_new
    ADD CONSTRAINT hosts_new_pkey PRIMARY KEY (org_id, id);


--
-- Name: hosts_p0 hosts_p0_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_p0
    ADD CONSTRAINT hosts_p0_pkey PRIMARY KEY (org_id, id);


--
-- Name: hosts_p1 hosts_p1_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_p1
    ADD CONSTRAINT hosts_p1_pkey PRIMARY KEY (org_id, id);


--
-- Name: hosts hosts_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts
    ADD CONSTRAINT hosts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY hbi.hosts REPLICA IDENTITY USING INDEX hosts_pkey;


--
-- Name: staleness staleness_org_id_key; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.staleness
    ADD CONSTRAINT staleness_org_id_key UNIQUE (org_id);


--
-- Name: staleness staleness_pkey; Type: CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.staleness
    ADD CONSTRAINT staleness_pkey PRIMARY KEY (id);

ALTER TABLE ONLY hbi.staleness REPLICA IDENTITY USING INDEX staleness_pkey;


--
-- Name: idx_hosts_groups_reverse; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_groups_reverse ON ONLY hbi.hosts_groups_new USING btree (org_id, group_id, host_id);


--
-- Name: hosts_groups_p0_org_id_group_id_host_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_groups_p0_org_id_group_id_host_id_idx ON hbi.hosts_groups_p0 USING btree (org_id, group_id, host_id);


--
-- Name: idx_hosts_groups_forward; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_groups_forward ON ONLY hbi.hosts_groups_new USING btree (org_id, host_id, group_id);


--
-- Name: hosts_groups_p0_org_id_host_id_group_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_groups_p0_org_id_host_id_group_id_idx ON hbi.hosts_groups_p0 USING btree (org_id, host_id, group_id);


--
-- Name: hosts_groups_p1_org_id_group_id_host_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_groups_p1_org_id_group_id_host_id_idx ON hbi.hosts_groups_p1 USING btree (org_id, group_id, host_id);


--
-- Name: hosts_groups_p1_org_id_host_id_group_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_groups_p1_org_id_host_id_group_id_idx ON hbi.hosts_groups_p1 USING btree (org_id, host_id, group_id);


--
-- Name: hosts_modified_on_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_modified_on_id ON hbi.hosts USING btree (modified_on DESC, id DESC);


--
-- Name: idx_hosts_sap_system; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_sap_system ON ONLY hbi.hosts_new USING btree ((((system_profile_facts ->> 'sap_system'::text))::boolean));


--
-- Name: hosts_p0_bool_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_bool_idx ON hbi.hosts_p0 USING btree ((((system_profile_facts ->> 'sap_system'::text))::boolean));


--
-- Name: idx_hosts_canonical_facts_gin; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_canonical_facts_gin ON ONLY hbi.hosts_new USING gin (canonical_facts jsonb_path_ops);


--
-- Name: hosts_p0_canonical_facts_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_canonical_facts_idx ON hbi.hosts_p0 USING gin (canonical_facts jsonb_path_ops);


--
-- Name: idx_hosts_host_type; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_host_type ON ONLY hbi.hosts_new USING btree (((system_profile_facts ->> 'host_type'::text)));


--
-- Name: hosts_p0_expr_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_expr_idx ON hbi.hosts_p0 USING btree (((system_profile_facts ->> 'host_type'::text)));


--
-- Name: idx_hosts_cf_insights_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_cf_insights_id ON ONLY hbi.hosts_new USING btree (((canonical_facts ->> 'insights_id'::text)));


--
-- Name: hosts_p0_expr_idx1; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_expr_idx1 ON hbi.hosts_p0 USING btree (((canonical_facts ->> 'insights_id'::text)));


--
-- Name: idx_hosts_cf_subscription_manager_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_cf_subscription_manager_id ON ONLY hbi.hosts_new USING btree (((canonical_facts ->> 'subscription_manager_id'::text)));


--
-- Name: hosts_p0_expr_idx2; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_expr_idx2 ON hbi.hosts_p0 USING btree (((canonical_facts ->> 'subscription_manager_id'::text)));


--
-- Name: idx_hosts_mssql; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_mssql ON ONLY hbi.hosts_new USING btree (((system_profile_facts ->> 'mssql'::text)));


--
-- Name: hosts_p0_expr_idx3; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_expr_idx3 ON hbi.hosts_p0 USING btree (((system_profile_facts ->> 'mssql'::text)));


--
-- Name: idx_hosts_ansible; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_ansible ON ONLY hbi.hosts_new USING btree (((system_profile_facts ->> 'ansible'::text)));


--
-- Name: hosts_p0_expr_idx4; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_expr_idx4 ON hbi.hosts_p0 USING btree (((system_profile_facts ->> 'ansible'::text)));


--
-- Name: idx_hosts_operating_system_multi; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_operating_system_multi ON ONLY hbi.hosts_new USING btree ((((system_profile_facts -> 'operating_system'::text) ->> 'name'::text)), ((((system_profile_facts -> 'operating_system'::text) ->> 'major'::text))::integer), ((((system_profile_facts -> 'operating_system'::text) ->> 'minor'::text))::integer), ((system_profile_facts ->> 'host_type'::text)), modified_on, org_id) WHERE ((system_profile_facts -> 'operating_system'::text) IS NOT NULL);


--
-- Name: hosts_p0_expr_int4_int41_expr1_modified_on_org_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_expr_int4_int41_expr1_modified_on_org_id_idx ON hbi.hosts_p0 USING btree ((((system_profile_facts -> 'operating_system'::text) ->> 'name'::text)), ((((system_profile_facts -> 'operating_system'::text) ->> 'major'::text))::integer), ((((system_profile_facts -> 'operating_system'::text) ->> 'minor'::text))::integer), ((system_profile_facts ->> 'host_type'::text)), modified_on, org_id) WHERE ((system_profile_facts -> 'operating_system'::text) IS NOT NULL);


--
-- Name: idx_hosts_groups_gin; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_groups_gin ON ONLY hbi.hosts_new USING gin (groups);


--
-- Name: hosts_p0_groups_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_groups_idx ON hbi.hosts_p0 USING gin (groups);


--
-- Name: idx_hosts_insights_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_insights_id ON ONLY hbi.hosts_new USING btree (insights_id);


--
-- Name: hosts_p0_insights_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_insights_id_idx ON hbi.hosts_p0 USING btree (insights_id);


--
-- Name: idx_hosts_modified_on_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_modified_on_id ON ONLY hbi.hosts_new USING btree (modified_on DESC, id DESC);


--
-- Name: hosts_p0_modified_on_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_modified_on_id_idx ON hbi.hosts_p0 USING btree (modified_on DESC, id DESC);


--
-- Name: idx_hosts_bootc_status; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_bootc_status ON ONLY hbi.hosts_new USING btree (org_id) WHERE ((((system_profile_facts -> 'bootc_status'::text) -> 'booted'::text) ->> 'image_digest'::text) IS NOT NULL);


--
-- Name: hosts_p0_org_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_org_id_idx ON hbi.hosts_p0 USING btree (org_id) WHERE ((((system_profile_facts -> 'bootc_status'::text) -> 'booted'::text) ->> 'image_digest'::text) IS NOT NULL);


--
-- Name: idx_hosts_host_type_modified_on_org_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_host_type_modified_on_org_id ON ONLY hbi.hosts_new USING btree (org_id, modified_on, ((system_profile_facts ->> 'host_type'::text)));


--
-- Name: hosts_p0_org_id_modified_on_expr_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_org_id_modified_on_expr_idx ON hbi.hosts_p0 USING btree (org_id, modified_on, ((system_profile_facts ->> 'host_type'::text)));


--
-- Name: idx_hosts_subscription_manager_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_hosts_subscription_manager_id ON ONLY hbi.hosts_new USING btree (subscription_manager_id);


--
-- Name: hosts_p0_subscription_manager_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p0_subscription_manager_id_idx ON hbi.hosts_p0 USING btree (subscription_manager_id);


--
-- Name: hosts_p1_bool_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_bool_idx ON hbi.hosts_p1 USING btree ((((system_profile_facts ->> 'sap_system'::text))::boolean));


--
-- Name: hosts_p1_canonical_facts_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_canonical_facts_idx ON hbi.hosts_p1 USING gin (canonical_facts jsonb_path_ops);


--
-- Name: hosts_p1_expr_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_expr_idx ON hbi.hosts_p1 USING btree (((system_profile_facts ->> 'host_type'::text)));


--
-- Name: hosts_p1_expr_idx1; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_expr_idx1 ON hbi.hosts_p1 USING btree (((canonical_facts ->> 'insights_id'::text)));


--
-- Name: hosts_p1_expr_idx2; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_expr_idx2 ON hbi.hosts_p1 USING btree (((canonical_facts ->> 'subscription_manager_id'::text)));


--
-- Name: hosts_p1_expr_idx3; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_expr_idx3 ON hbi.hosts_p1 USING btree (((system_profile_facts ->> 'mssql'::text)));


--
-- Name: hosts_p1_expr_idx4; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_expr_idx4 ON hbi.hosts_p1 USING btree (((system_profile_facts ->> 'ansible'::text)));


--
-- Name: hosts_p1_expr_int4_int41_expr1_modified_on_org_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_expr_int4_int41_expr1_modified_on_org_id_idx ON hbi.hosts_p1 USING btree ((((system_profile_facts -> 'operating_system'::text) ->> 'name'::text)), ((((system_profile_facts -> 'operating_system'::text) ->> 'major'::text))::integer), ((((system_profile_facts -> 'operating_system'::text) ->> 'minor'::text))::integer), ((system_profile_facts ->> 'host_type'::text)), modified_on, org_id) WHERE ((system_profile_facts -> 'operating_system'::text) IS NOT NULL);


--
-- Name: hosts_p1_groups_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_groups_idx ON hbi.hosts_p1 USING gin (groups);


--
-- Name: hosts_p1_insights_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_insights_id_idx ON hbi.hosts_p1 USING btree (insights_id);


--
-- Name: hosts_p1_modified_on_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_modified_on_id_idx ON hbi.hosts_p1 USING btree (modified_on DESC, id DESC);


--
-- Name: hosts_p1_org_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_org_id_idx ON hbi.hosts_p1 USING btree (org_id) WHERE ((((system_profile_facts -> 'bootc_status'::text) -> 'booted'::text) ->> 'image_digest'::text) IS NOT NULL);


--
-- Name: hosts_p1_org_id_modified_on_expr_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_org_id_modified_on_expr_idx ON hbi.hosts_p1 USING btree (org_id, modified_on, ((system_profile_facts ->> 'host_type'::text)));


--
-- Name: hosts_p1_subscription_manager_id_idx; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX hosts_p1_subscription_manager_id_idx ON hbi.hosts_p1 USING btree (subscription_manager_id);


--
-- Name: idx_groups_org_id_name_nocase; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE UNIQUE INDEX idx_groups_org_id_name_nocase ON hbi.groups USING btree (org_id, lower((name)::text));


--
-- Name: idx_host_type; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_host_type ON hbi.hosts USING btree (((system_profile_facts ->> 'host_type'::text)));


--
-- Name: idx_host_type_modified_on_org_id; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_host_type_modified_on_org_id ON hbi.hosts USING btree (org_id, modified_on, ((system_profile_facts ->> 'host_type'::text)));


--
-- Name: idx_operating_system_multi; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idx_operating_system_multi ON hbi.hosts USING btree ((((system_profile_facts -> 'operating_system'::text) ->> 'name'::text)), ((((system_profile_facts -> 'operating_system'::text) ->> 'major'::text))::integer), ((((system_profile_facts -> 'operating_system'::text) ->> 'minor'::text))::integer), ((system_profile_facts ->> 'host_type'::text)), modified_on, org_id) WHERE ((system_profile_facts -> 'operating_system'::text) IS NOT NULL);


--
-- Name: idxaccstaleorgid; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxaccstaleorgid ON hbi.staleness USING btree (org_id);


--
-- Name: idxansible; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxansible ON hbi.hosts USING btree (((system_profile_facts ->> 'ansible'::text)));


--
-- Name: idxbootc_status; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxbootc_status ON hbi.hosts USING btree (org_id) WHERE ((((system_profile_facts -> 'bootc_status'::text) -> 'booted'::text) ->> 'image_digest'::text) IS NOT NULL);


--
-- Name: idxgincanonicalfacts; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxgincanonicalfacts ON hbi.hosts USING gin (canonical_facts jsonb_path_ops);


--
-- Name: idxgrouporgid; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxgrouporgid ON hbi.groups USING btree (org_id);


--
-- Name: idxgroupshosts; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE UNIQUE INDEX idxgroupshosts ON hbi.hosts_groups USING btree (group_id, host_id);


--
-- Name: idxhostsgroups; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE UNIQUE INDEX idxhostsgroups ON hbi.hosts_groups USING btree (host_id, group_id);


--
-- Name: idxinsightsid; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxinsightsid ON hbi.hosts USING btree (((canonical_facts ->> 'insights_id'::text)));


--
-- Name: idxmssql; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxmssql ON hbi.hosts USING btree (((system_profile_facts ->> 'mssql'::text)));


--
-- Name: idxorgid; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxorgid ON hbi.hosts USING btree (org_id);


--
-- Name: idxorgidungrouped; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxorgidungrouped ON hbi.groups USING btree (org_id, ungrouped);


--
-- Name: idxsap_system; Type: INDEX; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

CREATE INDEX idxsap_system ON hbi.hosts USING btree ((((system_profile_facts ->> 'sap_system'::text))::boolean));


--
-- Name: hosts_groups_p0_org_id_group_id_host_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_groups_reverse ATTACH PARTITION hbi.hosts_groups_p0_org_id_group_id_host_id_idx;


--
-- Name: hosts_groups_p0_org_id_host_id_group_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_groups_forward ATTACH PARTITION hbi.hosts_groups_p0_org_id_host_id_group_id_idx;


--
-- Name: hosts_groups_p0_pkey; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.hosts_groups_new_pkey ATTACH PARTITION hbi.hosts_groups_p0_pkey;


--
-- Name: hosts_groups_p1_org_id_group_id_host_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_groups_reverse ATTACH PARTITION hbi.hosts_groups_p1_org_id_group_id_host_id_idx;


--
-- Name: hosts_groups_p1_org_id_host_id_group_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_groups_forward ATTACH PARTITION hbi.hosts_groups_p1_org_id_host_id_group_id_idx;


--
-- Name: hosts_groups_p1_pkey; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.hosts_groups_new_pkey ATTACH PARTITION hbi.hosts_groups_p1_pkey;


--
-- Name: hosts_p0_bool_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_sap_system ATTACH PARTITION hbi.hosts_p0_bool_idx;


--
-- Name: hosts_p0_canonical_facts_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_canonical_facts_gin ATTACH PARTITION hbi.hosts_p0_canonical_facts_idx;


--
-- Name: hosts_p0_expr_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_host_type ATTACH PARTITION hbi.hosts_p0_expr_idx;


--
-- Name: hosts_p0_expr_idx1; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_cf_insights_id ATTACH PARTITION hbi.hosts_p0_expr_idx1;


--
-- Name: hosts_p0_expr_idx2; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_cf_subscription_manager_id ATTACH PARTITION hbi.hosts_p0_expr_idx2;


--
-- Name: hosts_p0_expr_idx3; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_mssql ATTACH PARTITION hbi.hosts_p0_expr_idx3;


--
-- Name: hosts_p0_expr_idx4; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_ansible ATTACH PARTITION hbi.hosts_p0_expr_idx4;


--
-- Name: hosts_p0_expr_int4_int41_expr1_modified_on_org_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_operating_system_multi ATTACH PARTITION hbi.hosts_p0_expr_int4_int41_expr1_modified_on_org_id_idx;


--
-- Name: hosts_p0_groups_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_groups_gin ATTACH PARTITION hbi.hosts_p0_groups_idx;


--
-- Name: hosts_p0_insights_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_insights_id ATTACH PARTITION hbi.hosts_p0_insights_id_idx;


--
-- Name: hosts_p0_modified_on_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_modified_on_id ATTACH PARTITION hbi.hosts_p0_modified_on_id_idx;


--
-- Name: hosts_p0_org_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_bootc_status ATTACH PARTITION hbi.hosts_p0_org_id_idx;


--
-- Name: hosts_p0_org_id_modified_on_expr_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_host_type_modified_on_org_id ATTACH PARTITION hbi.hosts_p0_org_id_modified_on_expr_idx;


--
-- Name: hosts_p0_pkey; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.hosts_new_pkey ATTACH PARTITION hbi.hosts_p0_pkey;


--
-- Name: hosts_p0_subscription_manager_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_subscription_manager_id ATTACH PARTITION hbi.hosts_p0_subscription_manager_id_idx;


--
-- Name: hosts_p1_bool_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_sap_system ATTACH PARTITION hbi.hosts_p1_bool_idx;


--
-- Name: hosts_p1_canonical_facts_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_canonical_facts_gin ATTACH PARTITION hbi.hosts_p1_canonical_facts_idx;


--
-- Name: hosts_p1_expr_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_host_type ATTACH PARTITION hbi.hosts_p1_expr_idx;


--
-- Name: hosts_p1_expr_idx1; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_cf_insights_id ATTACH PARTITION hbi.hosts_p1_expr_idx1;


--
-- Name: hosts_p1_expr_idx2; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_cf_subscription_manager_id ATTACH PARTITION hbi.hosts_p1_expr_idx2;


--
-- Name: hosts_p1_expr_idx3; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_mssql ATTACH PARTITION hbi.hosts_p1_expr_idx3;


--
-- Name: hosts_p1_expr_idx4; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_ansible ATTACH PARTITION hbi.hosts_p1_expr_idx4;


--
-- Name: hosts_p1_expr_int4_int41_expr1_modified_on_org_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_operating_system_multi ATTACH PARTITION hbi.hosts_p1_expr_int4_int41_expr1_modified_on_org_id_idx;


--
-- Name: hosts_p1_groups_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_groups_gin ATTACH PARTITION hbi.hosts_p1_groups_idx;


--
-- Name: hosts_p1_insights_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_insights_id ATTACH PARTITION hbi.hosts_p1_insights_id_idx;


--
-- Name: hosts_p1_modified_on_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_modified_on_id ATTACH PARTITION hbi.hosts_p1_modified_on_id_idx;


--
-- Name: hosts_p1_org_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_bootc_status ATTACH PARTITION hbi.hosts_p1_org_id_idx;


--
-- Name: hosts_p1_org_id_modified_on_expr_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_host_type_modified_on_org_id ATTACH PARTITION hbi.hosts_p1_org_id_modified_on_expr_idx;


--
-- Name: hosts_p1_pkey; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.hosts_new_pkey ATTACH PARTITION hbi.hosts_p1_pkey;


--
-- Name: hosts_p1_subscription_manager_id_idx; Type: INDEX ATTACH; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER INDEX hbi.idx_hosts_subscription_manager_id ATTACH PARTITION hbi.hosts_p1_subscription_manager_id_idx;


--
-- Name: hosts_groups_new fk_hosts_groups_on_groups; Type: FK CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE hbi.hosts_groups_new
    ADD CONSTRAINT fk_hosts_groups_on_groups FOREIGN KEY (group_id) REFERENCES hbi.groups(id);


--
-- Name: hosts_groups_new fk_hosts_groups_on_hosts; Type: FK CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE hbi.hosts_groups_new
    ADD CONSTRAINT fk_hosts_groups_on_hosts FOREIGN KEY (org_id, host_id) REFERENCES hbi.hosts_new(org_id, id) ON DELETE CASCADE;


--
-- Name: hosts_groups hosts_groups_group_id_fkey; Type: FK CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups
    ADD CONSTRAINT hosts_groups_group_id_fkey FOREIGN KEY (group_id) REFERENCES hbi.groups(id);


--
-- Name: hosts_groups hosts_groups_host_id_fkey; Type: FK CONSTRAINT; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER TABLE ONLY hbi.hosts_groups
    ADD CONSTRAINT hosts_groups_host_id_fkey FOREIGN KEY (host_id) REFERENCES hbi.hosts(id);


--
-- Name: hbi_hosts_pub_v1_0_1; Type: PUBLICATION; Schema: -; Owner: VBzEIpnveJmm3TXi
--

CREATE PUBLICATION hbi_hosts_pub_v1_0_1 WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION hbi_hosts_pub_v1_0_1 OWNER TO "VBzEIpnveJmm3TXi";

--
-- Name: hbi_hosts_pub_v1_0_1 hosts; Type: PUBLICATION TABLE; Schema: hbi; Owner: VBzEIpnveJmm3TXi
--

ALTER PUBLICATION hbi_hosts_pub_v1_0_1 ADD TABLE ONLY hbi.hosts (id, account, display_name, created_on, modified_on, facts, canonical_facts, system_profile_facts, ansible_host, stale_timestamp, reporter, per_reporter_staleness, org_id, groups, tags_alt, last_check_in, stale_warning_timestamp, deletion_timestamp);

CREATE TABLE hbi.signal (id VARCHAR(255) PRIMARY KEY, type VARCHAR(255) NOT NULL, data VARCHAR(255) NULL);

--
-- PostgreSQL database dump complete
--

