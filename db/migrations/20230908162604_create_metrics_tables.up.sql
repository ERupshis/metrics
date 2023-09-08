CREATE SCHEMA IF NOT EXISTS metrics;

CREATE TABLE IF NOT EXISTS metrics.gauges (id TEXT PRIMARY KEY, value float8);
CREATE TABLE IF NOT EXISTS metrics.counters (id TEXT PRIMARY KEY, value int8);