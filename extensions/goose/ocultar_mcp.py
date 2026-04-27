#!/usr/bin/env python3
"""Ocultar PII Refinery — Goose MCP Extension (stdio transport).

Zero-egress guarantee: if Ocultar is unreachable this server returns an MCP
error and refuses to forward the raw text to the caller.
"""

import asyncio
import json
import os
import re

import httpx
import mcp.server.stdio
import mcp.types as types
from mcp.server import Server

OCULTAR_URL = os.environ.get("OCULTAR_URL", "http://localhost:8080").rstrip("/")
OCULTAR_API_KEY = os.environ.get("OCULTAR_API_KEY", "")

# Matches tokens like [EMAIL_9c8f7a1b] produced by the Ocultar tokenizer
_TOKEN_RE = re.compile(r"\[([A-Z_]+)_([a-f0-9]{8})\]")

app = Server("ocultar-pii")


@app.list_tools()
async def list_tools() -> list[types.Tool]:
    return [
        types.Tool(
            name="refine_text",
            description=(
                "Redact PII from text using the local Ocultar Refinery. "
                "Returns the cleaned text with all PII replaced by deterministic "
                "tokens (e.g. [EMAIL_9c8f7a1b]) and a map of each token to its "
                "PII type. All processing happens locally — no data leaves your "
                "infrastructure."
            ),
            inputSchema={
                "type": "object",
                "properties": {
                    "text": {
                        "type": "string",
                        "description": "Raw text that may contain PII.",
                    }
                },
                "required": ["text"],
            },
        )
    ]


@app.call_tool()
async def call_tool(name: str, arguments: dict) -> list[types.TextContent]:
    if name != "refine_text":
        raise ValueError(f"Unknown tool: {name}")

    text = arguments.get("text", "")
    if not text.strip():
        payload = json.dumps({"cleanText": "", "tokenMap": {}}, ensure_ascii=False)
        return [types.TextContent(type="text", text=payload)]

    headers: dict[str, str] = {}
    if OCULTAR_API_KEY:
        headers["Authorization"] = f"Bearer {OCULTAR_API_KEY}"

    try:
        async with httpx.AsyncClient(timeout=15.0) as client:
            response = await client.post(
                f"{OCULTAR_URL}/api/refine",
                content=text.encode("utf-8"),
                headers=headers,
            )
            response.raise_for_status()
    except httpx.ConnectError:
        raise RuntimeError(
            f"Cannot connect to Ocultar at {OCULTAR_URL}. "
            "Start the Refinery first: `docker compose up` or "
            "`go run ./services/refinery/cmd/main.go --serve 8080`. "
            "Raw text withheld to preserve zero-egress guarantee."
        )
    except httpx.TimeoutException:
        raise RuntimeError(
            f"Ocultar request timed out (15 s) at {OCULTAR_URL}. "
            "Raw text withheld to preserve zero-egress guarantee."
        )
    except httpx.HTTPStatusError as exc:
        raise RuntimeError(
            f"Ocultar returned HTTP {exc.response.status_code}: "
            f"{exc.response.text[:300]}"
        )

    data = response.json()
    clean_text: str = data.get("refined", "")

    # Build token map from the redacted output: token → PII type
    # e.g. {"[EMAIL_9c8f7a1b]": "EMAIL", "[SSN_3a1b2c4d]": "SSN"}
    token_map: dict[str, str] = {
        m.group(0): m.group(1) for m in _TOKEN_RE.finditer(clean_text)
    }

    payload = json.dumps(
        {"cleanText": clean_text, "tokenMap": token_map},
        ensure_ascii=False,
    )
    return [types.TextContent(type="text", text=payload)]


async def main() -> None:
    async with mcp.server.stdio.stdio_server() as (read_stream, write_stream):
        await app.run(
            read_stream,
            write_stream,
            app.create_initialization_options(),
        )


def main_sync() -> None:
    """Entrypoint for the ocultar-goose-mcp console script."""
    asyncio.run(main())


if __name__ == "__main__":
    main_sync()
