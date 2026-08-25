from __future__ import annotations

import argparse
import hashlib
import json
import shutil
import sqlite3
import tempfile
import time
import urllib.parse
import urllib.request
import zipfile
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Callable, Sequence

CAOP_URL = "https://geo2.dgterritorio.gov.pt/caop/CAOP_Continente_2025-gpkg.zip"
CAOP_SHA256 = "87cd67f4b1fbadf23d9324e6fb231ff05531e4db347af36ccc7c6cbabe3ecd1d"
SOURCE_VERSION = "2025"
RETRIEVAL_DATE = "2026-08-23"
NOMINATIM_URL = "https://nominatim.openstreetmap.org/search"
USER_AGENT = "JuntlyReferenceVerifier/1.0 (SourceSensei; one-time launch data verification)"


class ReferenceDataError(ValueError):
    pass


@dataclass(frozen=True)
class RequiredArea:
    kind: str
    external_code: str
    name: str
    parent_code: str | None
    table: str | None
    code_column: str | None
    name_column: str | None


@dataclass(frozen=True)
class AdministrativeArea:
    kind: str
    external_code: str
    name: str
    parent_code: str | None


@dataclass(frozen=True)
class FrozenLocality:
    slug: str
    name: str
    query: str
    parent_code: str
    osm_type: str
    osm_id: int


@dataclass(frozen=True)
class Locality:
    slug: str
    name: str
    parent_code: str
    source_element_id: str
    latitude: str
    longitude: str


REQUIRED_AREAS = (
    RequiredArea("country", "PT", "Portugal", None, None, None, None),
    RequiredArea("district", "05", "Castelo Branco", "PT", "cont_distritos", "dt", "distrito"),
    RequiredArea("municipality", "0502", "Castelo Branco", "05", "cont_municipios", "dtmn", "municipio"),
    RequiredArea("municipality", "0505", "Idanha-a-Nova", "05", "cont_municipios", "dtmn", "municipio"),
    RequiredArea("parish", "050205", "Castelo Branco", "0502", "cont_freguesias", "dtmnfr", "freguesia"),
    RequiredArea("parish", "050510", "Penha Garcia", "0505", "cont_freguesias", "dtmnfr", "freguesia"),
    RequiredArea(
        "parish",
        "050518",
        "União das freguesias de Idanha-a-Nova e Alcafozes",
        "0505",
        "cont_freguesias",
        "dtmnfr",
        "freguesia",
    ),
    RequiredArea(
        "parish",
        "050520",
        "União das freguesias de Monsanto e Idanha-a-Velha",
        "0505",
        "cont_freguesias",
        "dtmnfr",
        "freguesia",
    ),
    RequiredArea(
        "parish",
        "050521",
        "União das freguesias de Zebreira e Segura",
        "0505",
        "cont_freguesias",
        "dtmnfr",
        "freguesia",
    ),
)

FROZEN_LOCALITIES = (
    FrozenLocality(
        "castelo-branco",
        "Castelo Branco",
        "Castelo Branco, Castelo Branco, Portugal",
        "050205",
        "relation",
        5396187,
    ),
    FrozenLocality(
        "idanha-a-nova",
        "Idanha-a-Nova",
        "Idanha-a-Nova, Castelo Branco, Portugal",
        "050518",
        "relation",
        5395738,
    ),
    FrozenLocality(
        "zebreira",
        "Zebreira",
        "Zebreira, Idanha-a-Nova, Portugal",
        "050521",
        "node",
        440173641,
    ),
    FrozenLocality(
        "penha-garcia",
        "Penha Garcia",
        "Penha Garcia, Idanha-a-Nova, Portugal",
        "050510",
        "relation",
        5431477,
    ),
    FrozenLocality(
        "monsanto",
        "Monsanto",
        "Monsanto, Idanha-a-Nova, Portugal",
        "050520",
        "node",
        371426674,
    ),
)

CATEGORY_SEED = (
    (
        "home-repairs",
        None,
        10,
        ("Reparações domésticas", "Home repairs", "Reparaciones del hogar"),
    ),
    ("plumbing", "home-repairs", 10, ("Canalização", "Plumbing", "Fontanería")),
    ("electrical-work", "home-repairs", 20, ("Eletricidade", "Electrical work", "Electricidad")),
    ("construction", "home-repairs", 30, ("Construção", "Construction", "Construcción")),
    ("small-repairs", "home-repairs", 40, ("Pequenas reparações", "Small repairs", "Pequeñas reparaciones")),
    ("home-and-garden", None, 20, ("Casa e jardim", "Home and garden", "Hogar y jardín")),
    ("cleaning", "home-and-garden", 10, ("Limpeza", "Cleaning", "Limpieza")),
    ("gardening", "home-and-garden", 20, ("Jardinagem", "Gardening", "Jardinería")),
    ("rural-and-transport", None, 30, ("Serviços rurais e transporte", "Rural services and transport", "Servicios rurales y transporte")),
    ("agricultural-assistance", "rural-and-transport", 10, ("Apoio agrícola", "Agricultural assistance", "Ayuda agrícola")),
    ("transport", "rural-and-transport", 20, ("Transporte", "Transport", "Transporte")),
    ("care-and-learning", None, 40, ("Cuidados e aprendizagem", "Care and learning", "Cuidados y aprendizaje")),
    ("elderly-assistance", "care-and-learning", 10, ("Apoio a idosos", "Elderly assistance", "Ayuda a mayores")),
    ("animal-care", "care-and-learning", 20, ("Cuidados de animais", "Animal care", "Cuidado de animales")),
    ("private-lessons", "care-and-learning", 30, ("Aulas particulares", "Private lessons", "Clases particulares")),
    ("food-and-technology", None, 50, ("Alimentação e tecnologia", "Food and technology", "Alimentación y tecnología")),
    ("meal-preparation", "food-and-technology", 10, ("Preparação de refeições", "Meal preparation", "Preparación de comidas")),
    ("computer-repair", "food-and-technology", 20, ("Reparação de computadores", "Computer repair", "Reparación de ordenadores")),
)

LANGUAGE_SEED = (
    ("pt-PT", 10, ("Português", "Portuguese", "Portugués")),
    ("en", 20, ("Inglês", "English", "Inglés")),
    ("es", 30, ("Espanhol", "Spanish", "Español")),
)

LOCALES = ("pt-PT", "en", "es")


def verify_caop_archive(path: Path, expected_sha256: str = CAOP_SHA256) -> None:
    digest = hashlib.sha256()
    with path.open("rb") as source:
        for chunk in iter(lambda: source.read(1024 * 1024), b""):
            digest.update(chunk)
    if digest.hexdigest() != expected_sha256:
        raise ReferenceDataError("CAOP checksum mismatch")


def extract_administrative_rows(
    archive: Path, required: Sequence[RequiredArea]
) -> list[AdministrativeArea]:
    with tempfile.TemporaryDirectory(prefix="juntly-caop-extract-") as directory:
        root = Path(directory)
        with zipfile.ZipFile(archive) as zipped:
            packages = [name for name in zipped.namelist() if name.lower().endswith(".gpkg")]
            if len(packages) != 1:
                raise ReferenceDataError("CAOP archive must contain exactly one GeoPackage")
            zipped.extract(packages[0], root)
        connection = sqlite3.connect(root / packages[0])
        try:
            rows: list[AdministrativeArea] = []
            for area in required:
                if area.table is None:
                    rows.append(
                        AdministrativeArea(
                            area.kind, area.external_code, area.name, area.parent_code
                        )
                    )
                    continue
                table = _allowlisted_identifier(area.table)
                code_column = _allowlisted_identifier(area.code_column or "")
                name_column = _allowlisted_identifier(area.name_column or "")
                matches = connection.execute(
                    f'SELECT "{code_column}", "{name_column}" FROM "{table}" '
                    f'WHERE "{code_column}" = ? AND "{name_column}" = ?',
                    (area.external_code, area.name),
                ).fetchall()
                if len(matches) != 1:
                    raise ReferenceDataError(
                        f"required administrative row {area.external_code} must exist exactly once"
                    )
                rows.append(
                    AdministrativeArea(
                        area.kind, area.external_code, area.name, area.parent_code
                    )
                )
            return rows
        finally:
            connection.close()


def resolve_localities(
    resolve: Callable[[str], list[dict[str, object]]],
    sleep: Callable[[float], None] = time.sleep,
) -> list[Locality]:
    localities: list[Locality] = []
    for index, frozen in enumerate(FROZEN_LOCALITIES):
        if index:
            sleep(1.05)
        results = resolve(frozen.query)
        matches = [
            item
            for item in results
            if item.get("osm_type") == frozen.osm_type
            and item.get("osm_id") == frozen.osm_id
        ]
        if len(matches) != 1:
            raise ReferenceDataError(
                f"frozen locality {frozen.slug} must resolve exactly once"
            )
        result = matches[0]
        latitude = _coordinate(result.get("lat"), -90, 90, "latitude")
        longitude = _coordinate(result.get("lon"), -180, 180, "longitude")
        prefix = "N" if frozen.osm_type == "node" else "R"
        localities.append(
            Locality(
                slug=frozen.slug,
                name=frozen.name,
                parent_code=frozen.parent_code,
                source_element_id=f"{prefix}{frozen.osm_id}",
                latitude=latitude,
                longitude=longitude,
            )
        )
    return sorted(localities, key=lambda item: item.slug)


def build_manifest(
    caop_zip: Path,
    resolve: Callable[[str], list[dict[str, object]]],
    sleep: Callable[[float], None] = time.sleep,
    expected_sha256: str = CAOP_SHA256,
) -> dict[str, object]:
    verify_caop_archive(caop_zip, expected_sha256)
    areas = extract_administrative_rows(caop_zip, REQUIRED_AREAS)
    localities = resolve_localities(resolve, sleep)
    return {
        "version": 1,
        "retrievalDate": RETRIEVAL_DATE,
        "source": {
            "caop": {
                "url": CAOP_URL,
                "version": SOURCE_VERSION,
                "sha256": expected_sha256,
                "layers": ["cont_distritos", "cont_municipios", "cont_freguesias"],
            },
            "openStreetMap": {
                "nominatimPolicy": "https://operations.osmfoundation.org/policies/nominatim",
                "license": "https://www.openstreetmap.org/copyright",
            },
        },
        "attribution": {
            "text": "© OpenStreetMap contributors",
            "url": "https://www.openstreetmap.org/copyright",
        },
        "supportedLocales": list(LOCALES),
        "administrativeAreas": [asdict(area) for area in areas],
        "localities": [asdict(locality) for locality in localities],
        "categories": [
            {
                "slug": slug,
                "parentSlug": parent,
                "sortOrder": order,
                "translations": dict(zip(LOCALES, names, strict=True)),
            }
            for slug, parent, order, names in CATEGORY_SEED
        ],
        "languages": [
            {
                "code": code,
                "sortOrder": order,
                "translations": dict(zip(LOCALES, names, strict=True)),
            }
            for code, order, names in LANGUAGE_SEED
        ],
    }


def resolve_nominatim(query: str) -> list[dict[str, object]]:
    url = NOMINATIM_URL + "?" + urllib.parse.urlencode(
        {"q": query, "format": "jsonv2", "limit": "3", "countrycodes": "pt"}
    )
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(request, timeout=30) as response:
        payload = json.load(response)
    if not isinstance(payload, list):
        raise ReferenceDataError("Nominatim response must be a list")
    return payload


def download_caop(destination: Path) -> None:
    request = urllib.request.Request(CAOP_URL, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(request, timeout=120) as response, destination.open(
        "wb"
    ) as output:
        shutil.copyfileobj(response, output)


def _allowlisted_identifier(value: str) -> str:
    allowed = {
        "cont_distritos",
        "cont_municipios",
        "cont_freguesias",
        "dt",
        "dtmn",
        "dtmnfr",
        "distrito",
        "municipio",
        "freguesia",
    }
    if value not in allowed:
        raise ReferenceDataError("unexpected CAOP identifier")
    return value


def _coordinate(value: object, minimum: int, maximum: int, label: str) -> str:
    if not isinstance(value, str):
        raise ReferenceDataError(f"{label} must be a string")
    try:
        numeric = float(value)
    except ValueError as error:
        raise ReferenceDataError(f"invalid {label}") from error
    if not minimum <= numeric <= maximum:
        raise ReferenceDataError(f"invalid {label}")
    return value


def _write_json(path: Path, manifest: dict[str, object]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )


def main() -> int:
    parser = argparse.ArgumentParser()
    mode = parser.add_mutually_exclusive_group(required=True)
    mode.add_argument("--write", type=Path)
    mode.add_argument("--verify", type=Path)
    args = parser.parse_args()

    temporary_root = Path(tempfile.mkdtemp(prefix="juntly-launch-reference-"))
    try:
        archive = temporary_root / "caop-2025.zip"
        download_caop(archive)
        manifest = build_manifest(archive, resolve_nominatim)
        target: Path = args.write or args.verify
        if args.write:
            _write_json(target, manifest)
            return 0
        expected = json.loads(target.read_text(encoding="utf-8"))
        if manifest != expected:
            raise ReferenceDataError("checked-in launch manifest does not match sources")
        return 0
    finally:
        shutil.rmtree(temporary_root, ignore_errors=True)


if __name__ == "__main__":
    raise SystemExit(main())
