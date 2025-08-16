## LakeLens - Real Time Metadata Viewer
A platform to explore, audit and analyze your data lake tables Iceberg, Delta and Hudi.

---
> ### LakeLens is under active development and is yet not ready for 1.0 launch.
---

### Features
-   Automatic Table Detection - Detects tables from any given data lake path, at any depth, using both standard and intelligent pattern matching
-   Full Metadata Exploration - Explore complete metadata for any table type.
-   Schema and Metrics Inspection - Inspect schemas, track metrics and compare with previous versions.
-   Snapshot Details - Get in-depth details about snapshots, from raw .avro or other files too, not just .json .
-   Manifest Insights - Access fine-grained metrics from manifests, including per-row, per-column statistics.
-   Files and Data - Visualize data flow and table changes with clear file lineages.
-   Multi-format Support - Works with Parquet, Avro, Json, etc files too.
-   Frontend and Productivity - No-nonsense, interactive UI with drag-and-drop and dynamic graphs to build your ideal view.
-   Performance centric - Optimized to handle tables of any size and age, from millions to hundreds of millions of rows, with consistently fast performance.

---

### Currently Working On
-   Cache invalidation ;) It's expensive/tricky to determine updates in virtual file systems (like S3) where event propagation doesn't occur.
-   Frontend for all table types (it has to be customized for each type to maintain familarity for the user)
-   Extending to other data lake providers like MinIO and Azure, cause they are AWS based, GCP later on. 
-   Things like activity, metrics, auxiliary actionables, etc.
-   Internal algorithm/flow optimizations to allow even deeper detection of tables (it's currently limited to a certain depth because it's expensive for the provider as well as to me)

---

### Next Features
-   Querying capabilities, without having to write scripts, possibly using SQL or NLP.
-   Selective data fetching capabilites directly from tables.
-   Clean formatted report generation and export capabilites to allow for better audits.
-   AI powered warnings, suggestions and health analysis, there are privacy issues though.

