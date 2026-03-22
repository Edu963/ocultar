#!/usr/bin/env python3
# tools/scripts/scripts/manifest_executor.py
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Manifest-Driven Build Executor
# ─────────────────────────────────────────────────────────────────────────────

import yaml
import subprocess
import os
import shutil
import sys
import json

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../../.."))
DIST_DIR = os.path.join(REPO_ROOT, "dist")
MANIFEST_PATH = os.path.join(REPO_ROOT, "dist.manifest.yaml")

def patch_binary_key(filepath):
    """Replaces 64-zero dummy keys with a 64-byte placeholder in compiled binaries."""
    if not os.path.exists(filepath):
        return
    try:
        with open(filepath, "rb") as f:
            data = f.read()
            
        dummy_key = b"0" * 64
        if dummy_key in data:
            # 64-byte exact replacement
            placeholder = b"OCU_MASTER_KEY_PLACEHOLDER______________________________________"
            data = data.replace(dummy_key, placeholder)
            with open(filepath, "wb") as f:
                f.write(data)
            print(f"[*] Patched dummy key in binary: {filepath}")
    except Exception as e:
        print(f"[!] Warning: failed to patch binary {filepath}: {e}")

def log(msg, color="blue"):
    colors = {"blue": "\033[0;34m", "green": "\033[0;32m", "red": "\033[0;31m", "nc": "\033[0m"}
    print(f"{colors.get(color, colors['nc'])}{msg}{colors['nc']}")

def run(cmd, cwd=REPO_ROOT, optional=False):
    log(f"Executing: {cmd}", color="nc")
    result = subprocess.run(cmd, shell=True, cwd=cwd)
    if result.returncode != 0:
        if optional:
            log(f"Optional command failed with exit code {result.returncode}. Skipping.", color="blue")
        else:
            log(f"Command failed with exit code {result.returncode}", color="red")
            sys.exit(1)

def build_component(comp, target_dir):
    c_type = comp.get("type")
    src = os.path.join(REPO_ROOT, comp.get("src"))
    dest = os.path.join(target_dir, comp.get("dest"))
    optional = comp.get("optional", False)

    if c_type == "go-binary":
        log(f"Building Go binary: {comp['name']}")
        # Use -trimpath to remove local absolute paths from the binary
        run(f"go build -trimpath -o {dest} {src}", optional=optional)
        patch_binary_key(dest)
    
    elif c_type == "npm-build":
        log(f"Building NPM package: {comp['name']}")
        run("npm install && npm run build", cwd=src)
        build_dist = os.path.join(src, "dist")
        if os.path.exists(build_dist):
            shutil.copytree(build_dist, dest, dirs_exist_ok=True)
    
    elif c_type == "file":
        log(f"Copying file: {comp['name']}")
        os.makedirs(os.path.dirname(dest), exist_ok=True)
        shutil.copy2(src, dest)
    
    elif c_type == "directory":
        log(f"Copying directory: {comp['name']}")
        os.makedirs(dest, exist_ok=True)
        include = comp.get("include", [])
        for f in include:
            shutil.copy2(os.path.join(src, f), os.path.join(dest, f))

def compress_distribution(dist_key, dist_cfg, target_dir):
    format = dist_cfg.get("format")
    name = dist_cfg.get("name", dist_key)
    output_base = os.path.join(DIST_DIR, name)
    
    log(f"Compressing distribution {dist_key} to {format}...")
    if format == "zip":
        shutil.make_archive(output_base, 'zip', target_dir)
    elif format == "tar.gz":
        shutil.make_archive(output_base, 'gztar', target_dir)
        # shutil.make_archive for gztar creates .tar.gz
    else:
        log(f"Warning: Unknown format {format} for {dist_key}", color="red")

def generate_sbom(target_dir):
    log("Generating SBOM...")
    # Go SBOM
    res = subprocess.run("go list -json -m all", shell=True, capture_output=True, text=True, cwd=REPO_ROOT)
    if res.returncode == 0:
        # Scrub internal paths from SBOM
        clean_res = res.stdout.replace(REPO_ROOT, ".")
        with open(os.path.join(target_dir, "sbom_go.json"), "w") as f:
            f.write(clean_res)
    
    # NPM SBOM (simulated for now)
    web_dir = os.path.join(REPO_ROOT, "apps/web")
    if os.path.exists(web_dir):
        res = subprocess.run("npm list --json", shell=True, capture_output=True, text=True, cwd=web_dir)
        if res.returncode == 0:
            # Scrub internal paths from NPM SBOM
            clean_res = res.stdout.replace(REPO_ROOT, ".")
            with open(os.path.join(target_dir, "sbom_npm.json"), "w") as f:
                f.write(clean_res)

def main():
    if not os.path.exists(MANIFEST_PATH):
        log(f"Error: Manifest not found at {MANIFEST_PATH}", color="red")
        sys.exit(1)

    with open(MANIFEST_PATH, "r") as f:
        manifest = yaml.safe_load(f)

    log(f"Starting Ocultar Build (Manifest v{manifest.get('version', 'unknown')})")

    # Clean dist
    if os.path.exists(DIST_DIR):
        shutil.rmtree(DIST_DIR)
    os.makedirs(DIST_DIR)

    dists = manifest.get("distributions", {})
    
    for dist_key, dist_cfg in dists.items():
        log(f"\nProcessing Distribution: {dist_key}", color="green")
        target_dir = os.path.join(DIST_DIR, dist_key)
        os.makedirs(target_dir, exist_ok=True)

        # Handle extends (shallow one-level for now)
        base_components = []
        if "extends" in dist_cfg:
            base_dist = dists.get(dist_cfg["extends"])
            if base_dist:
                base_components = base_dist.get("components", [])

        # Build all components
        for comp in base_components + dist_cfg.get("components", []):
            build_component(comp, target_dir)
        
        # Generate SBOM if security flags are present
        if manifest.get("security", {}).get("sanitization"):
            generate_sbom(target_dir)

        # NEW: Compress the distribution
        compress_distribution(dist_key, dist_cfg, target_dir)

if __name__ == "__main__":
    main()
