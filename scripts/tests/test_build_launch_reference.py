from __future__ import annotations

import hashlib
import json
import sqlite3
import tempfile
import unittest
import zipfile
from pathlib import Path

from scripts.build_launch_reference import (
    CAOP_SHA256,
    REQUIRED_AREAS,
    ReferenceDataError,
    build_manifest,
    extract_administrative_rows,
    resolve_localities,
    verify_caop_archive,
)


class LaunchReferenceBuilderTests(unittest.TestCase):
    def test_rejects_caop_checksum_mismatch(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            archive = self._create_archive(Path(directory))

            with self.assertRaisesRegex(ReferenceDataError, "checksum"):
                verify_caop_archive(archive, "0" * 64)

    def test_extracts_each_required_administrative_row_once_in_stable_order(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            archive = self._create_archive(Path(directory))

            rows = extract_administrative_rows(archive, REQUIRED_AREAS)

        self.assertEqual(
            [(row.kind, row.external_code, row.name, row.parent_code) for row in rows],
            [
                ("country", "PT", "Portugal", None),
                ("district", "05", "Castelo Branco", "PT"),
                ("municipality", "0502", "Castelo Branco", "05"),
                ("municipality", "0505", "Idanha-a-Nova", "05"),
                ("parish", "050205", "Castelo Branco", "0502"),
                ("parish", "050510", "Penha Garcia", "0505"),
                (
                    "parish",
                    "050518",
                    "União das freguesias de Idanha-a-Nova e Alcafozes",
                    "0505",
                ),
                (
                    "parish",
                    "050520",
                    "União das freguesias de Monsanto e Idanha-a-Velha",
                    "0505",
                ),
                (
                    "parish",
                    "050521",
                    "União das freguesias de Zebreira e Segura",
                    "0505",
                ),
            ],
        )

    def test_rejects_missing_or_duplicate_administrative_rows(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            missing = self._create_archive(root / "missing", omit_code="050510")
            duplicate = self._create_archive(root / "duplicate", duplicate_code="050510")

            with self.assertRaisesRegex(ReferenceDataError, "exactly once"):
                extract_administrative_rows(missing, REQUIRED_AREAS)
            with self.assertRaisesRegex(ReferenceDataError, "exactly once"):
                extract_administrative_rows(duplicate, REQUIRED_AREAS)

    def test_resolves_only_frozen_osm_elements_and_spaces_requests(self) -> None:
        calls: list[str] = []
        sleeps: list[float] = []

        def resolver(query: str) -> list[dict[str, object]]:
            calls.append(query)
            frozen = {
                "Castelo Branco, Castelo Branco, Portugal": (
                    "relation",
                    5396187,
                    "39.8266322",
                    "-7.4919318",
                ),
                "Idanha-a-Nova, Castelo Branco, Portugal": (
                    "relation",
                    5395738,
                    "39.9260883",
                    "-7.2436356",
                ),
                "Zebreira, Idanha-a-Nova, Portugal": (
                    "node",
                    440173641,
                    "39.8455920",
                    "-7.0703366",
                ),
                "Penha Garcia, Idanha-a-Nova, Portugal": (
                    "relation",
                    5431477,
                    "40.0422569",
                    "-7.0163521",
                ),
                "Monsanto, Idanha-a-Nova, Portugal": (
                    "node",
                    371426674,
                    "40.0387510",
                    "-7.1151133",
                ),
            }
            osm_type, osm_id, latitude, longitude = frozen[query]
            return [
                {
                    "osm_type": "node",
                    "osm_id": 1,
                    "lat": "0",
                    "lon": "0",
                    "display_name": "unreviewed candidate",
                },
                {
                    "osm_type": osm_type,
                    "osm_id": osm_id,
                    "lat": latitude,
                    "lon": longitude,
                    "display_name": query,
                },
            ]

        localities = resolve_localities(resolver, sleeps.append)

        self.assertEqual(len(calls), 5)
        self.assertEqual(sleeps, [1.05, 1.05, 1.05, 1.05])
        self.assertEqual(
            [(item.slug, item.source_element_id) for item in localities],
            [
                ("castelo-branco", "R5396187"),
                ("idanha-a-nova", "R5395738"),
                ("monsanto", "N371426674"),
                ("penha-garcia", "R5431477"),
                ("zebreira", "N440173641"),
            ],
        )

    def test_build_manifest_is_deterministic_and_excludes_raw_source_data(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            archive = self._create_archive(Path(directory))
            digest = hashlib.sha256(archive.read_bytes()).hexdigest()

            def resolver(query: str) -> list[dict[str, object]]:
                by_query = {
                    "Castelo Branco, Castelo Branco, Portugal": ("relation", 5396187, "39.8266322", "-7.4919318"),
                    "Idanha-a-Nova, Castelo Branco, Portugal": ("relation", 5395738, "39.9260883", "-7.2436356"),
                    "Zebreira, Idanha-a-Nova, Portugal": ("node", 440173641, "39.8455920", "-7.0703366"),
                    "Penha Garcia, Idanha-a-Nova, Portugal": ("relation", 5431477, "40.0422569", "-7.0163521"),
                    "Monsanto, Idanha-a-Nova, Portugal": ("node", 371426674, "40.0387510", "-7.1151133"),
                }
                osm_type, osm_id, latitude, longitude = by_query[query]
                return [{"osm_type": osm_type, "osm_id": osm_id, "lat": latitude, "lon": longitude}]

            first = build_manifest(archive, resolver, lambda _: None, digest)
            second = build_manifest(archive, resolver, lambda _: None, digest)

        self.assertEqual(first, second)
        encoded = json.dumps(first, sort_keys=True)
        self.assertNotIn("postal", encoded.lower())
        self.assertNotIn("display_name", encoded)
        self.assertNotIn("raw", encoded.lower())
        self.assertEqual(first["source"]["caop"]["sha256"], digest)
        self.assertEqual(len(first["localities"]), 5)
        self.assertEqual(first["attribution"]["text"], "© OpenStreetMap contributors")

    def _create_archive(
        self,
        root: Path,
        *,
        omit_code: str | None = None,
        duplicate_code: str | None = None,
    ) -> Path:
        root.mkdir(parents=True, exist_ok=True)
        package = root / "source.gpkg"
        connection = sqlite3.connect(package)
        try:
            connection.execute("create table cont_distritos (dt text, distrito text)")
            connection.execute(
                "create table cont_municipios (dtmn text, municipio text, distrito_ilha text)"
            )
            connection.execute(
                "create table cont_freguesias (dtmnfr text, freguesia text, municipio text, distrito_ilha text)"
            )
            connection.execute(
                "insert into cont_distritos values (?, ?)", ("05", "Castelo Branco")
            )
            connection.executemany(
                "insert into cont_municipios values (?, ?, ?)",
                (("0502", "Castelo Branco", "Castelo Branco"), ("0505", "Idanha-a-Nova", "Castelo Branco")),
            )
            rows = [
                ("050205", "Castelo Branco", "Castelo Branco", "Castelo Branco"),
                ("050510", "Penha Garcia", "Idanha-a-Nova", "Castelo Branco"),
                ("050518", "União das freguesias de Idanha-a-Nova e Alcafozes", "Idanha-a-Nova", "Castelo Branco"),
                ("050520", "União das freguesias de Monsanto e Idanha-a-Velha", "Idanha-a-Nova", "Castelo Branco"),
                ("050521", "União das freguesias de Zebreira e Segura", "Idanha-a-Nova", "Castelo Branco"),
            ]
            connection.executemany(
                "insert into cont_freguesias values (?, ?, ?, ?)",
                [row for row in rows if row[0] != omit_code],
            )
            if duplicate_code:
                duplicate = next(row for row in rows if row[0] == duplicate_code)
                connection.execute(
                    "insert into cont_freguesias values (?, ?, ?, ?)", duplicate
                )
            connection.commit()
        finally:
            connection.close()

        archive = root / "source.zip"
        with zipfile.ZipFile(archive, "w") as zipped:
            zipped.write(package, arcname="source.gpkg")
        return archive


if __name__ == "__main__":
    unittest.main()
