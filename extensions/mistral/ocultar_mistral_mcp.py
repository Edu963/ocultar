#!/usr/bin/env python3
"""Ocultar PII Refinery — Mistral Le Chat MCP Extension (stdio transport).

Two tools:
  refine_text   — redacts PII before sending text to Mistral Le Chat
  reveal_tokens — de-tokenizes specific tokens back to plaintext (auditor-only)

Zero-egress guarantee: if Ocultar is unreachable both tools return an MCP error
and refuse to forward raw text or vault data to the caller.
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
OCULTAR_AUDITOR_TOKEN = os.environ.get("OCULTAR_AUDITOR_TOKEN", "")

# Matches tokens like [EMAIL_9c8f7a1b] produced by the Ocultar tokenizer
_TOKEN_RE = re.compile(r"\[([A-Z_]+)_([a-f0-9]{8})\]")

app = Server("ocultar-pii")


def _auth_headers() -> dict[str, str]:
    if OCULTAR_API_KEY:
        return {"Authorization": f"Bearer {OCULTAR_API_KEY}"}
    return {}


def _auditor_headers() -> dict[str, str]:
    if OCULTAR_AUDITOR_TOKEN:
        return {"Authorization": f"Bearer {OCULTAR_AUDITOR_TOKEN}"}
    return {}


def _connection_error(endpoint: str) -> RuntimeError:
    return RuntimeError(
        f"Cannot connect to Ocultar at {OCULTAR_URL}{endpoint}. "
        "Start the Refinery first: `docker compose up` or "
        "`go run ./services/refinery/cmd/main.go --serve 8080`. "
        "Raw data withheld to preserve zero-egress guarantee."
    )


@app.list_tools()
async def list_tools() -> list[types.Tool]:
    return [
        types.Tool(
            name="refine_text",
            title="Redact PII from text",
            description=(
                "Redact PII from text using the local Ocultar Refinery before "
                "sending it to Mistral Le Chat or any other AI model. Returns the "
                "cleaned text with all PII replaced by deterministic tokens "
                "(e.g. [EMAIL_9c8f7a1b]) and a map of each token to its PII type. "
                "All processing is local — no data leaves your infrastructure. "
                "Use this tool before processing any text that may contain names, "
                "emails, phone numbers, IBANs, SIRET/SIREN numbers, credit cards, "
                "or addresses. Optimised for French and EU regulatory requirements "
                "(RGPD, CNIL, DSP2)."
            ),
            readOnlyHint=True,
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
        ),
        types.Tool(
            name="reveal_tokens",
            title="Reveal original PII from tokens",
            description=(
                "De-tokenize Ocultar PII tokens back to their original plaintext "
                "values. Requires OCULTAR_AUDITOR_TOKEN to be set — this is an "
                "auditor-only operation that is logged in the immutable audit trail. "
                "Use only when the authorized caller explicitly needs to retrieve "
                "the original PII values after AI processing is complete."
            ),
            readOnlyHint=True,
            inputSchema={
                "type": "object",
                "properties": {
                    "tokens": {
                        "type": "array",
                        "items": {"type": "string"},
                        "description": (
                            "List of Ocultar tokens to de-tokenize, "
                            "e.g. ['[EMAIL_9c8f7a1b]', '[IBAN_3a1b2c4d]']."
                        ),
                    }
                },
                "required": ["tokens"],
            },
        ),
    ]


@app.call_tool()
async def call_tool(name: str, arguments: dict) -> list[types.TextContent]:
    if name == "refine_text":
        return await _refine_text(arguments)
    if name == "reveal_tokens":
        return await _reveal_tokens(arguments)
    raise ValueError(f"Unknown tool: {name}")


async def _refine_text(arguments: dict) -> list[types.TextContent]:
    text = arguments.get("text", "")
    if not text.strip():
        return [types.TextContent(type="text", text=json.dumps({"cleanText": "", "tokenMap": {}}))]

    try:
        async with httpx.AsyncClient(timeout=15.0) as client:
            response = await client.post(
                f"{OCULTAR_URL}/api/refine",
                content=text.encode("utf-8"),
                headers=_auth_headers(),
            )
            response.raise_for_status()
    except httpx.ConnectError:
        raise _connection_error("/api/refine")
    except httpx.TimeoutException:
        raise RuntimeError(
            "Ocultar /api/refine timed out (15 s). "
            "Raw text withheld to preserve zero-egress guarantee."
        )
    except httpx.HTTPStatusError as exc:
        raise RuntimeError(
            f"Ocultar /api/refine returned HTTP {exc.response.status_code}: "
            f"{exc.response.text[:300]}"
        )

    data = response.json()
    clean_text: str = data.get("refined", "")
    token_map: dict[str, str] = {
        m.group(0): m.group(1) for m in _TOKEN_RE.finditer(clean_text)
    }

    payload = json.dumps(
        {"cleanText": clean_text, "tokenMap": token_map},
        ensure_ascii=False,
    )
    return [types.TextContent(type="text", text=payload)]


async def _reveal_tokens(arguments: dict) -> list[types.TextContent]:
    tokens: list[str] = arguments.get("tokens", [])
    if not tokens:
        return [types.TextContent(type="text", text=json.dumps({"results": {}}))]

    if not OCULTAR_AUDITOR_TOKEN:
        raise RuntimeError(
            "OCULTAR_AUDITOR_TOKEN is not set. "
            "Token reveal is an auditor-only operation. "
            "Set the environment variable to enable it."
        )

    try:
        async with httpx.AsyncClient(timeout=15.0) as client:
            response = await client.post(
                f"{OCULTAR_URL}/api/reveal",
                json={"tokens": tokens},
                headers=_auditor_headers(),
            )
            response.raise_for_status()
    except httpx.ConnectError:
        raise _connection_error("/api/reveal")
    except httpx.TimeoutException:
        raise RuntimeError(
            "Ocultar /api/reveal timed out (15 s). "
            "Token data withheld to preserve zero-egress guarantee."
        )
    except httpx.HTTPStatusError as exc:
        if exc.response.status_code in (401, 403):
            raise RuntimeError(
                "Ocultar rejected the auditor token (401/403). "
                "Check OCULTAR_AUDITOR_TOKEN matches OCU_AUDITOR_TOKEN on the server."
            )
        raise RuntimeError(
            f"Ocultar /api/reveal returned HTTP {exc.response.status_code}: "
            f"{exc.response.text[:300]}"
        )

    data = response.json()
    return [types.TextContent(type="text", text=json.dumps(data, ensure_ascii=False))]


async def main() -> None:
    async with mcp.server.stdio.stdio_server() as (read_stream, write_stream):
        await app.run(
            read_stream,
            write_stream,
            app.create_initialization_options(),
        )


def main_sync() -> None:
    """Entrypoint for the ocultar-mistral-mcp console script."""
    asyncio.run(main())


if __name__ == "__main__":
    main_sync()
