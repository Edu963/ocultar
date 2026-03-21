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
        run(f"go build -o {dest} {src}", optional=optional)
    
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

def generate_sbom(target_dir):
    log("Generating SBOM...")
    # Go SBOM
    res = subprocess.run("go list -json -m all", shell=True, capture_output=True, text=True, cwd=REPO_ROOT)
    if res.returncode == 0:
        with open(os.path.join(target_dir, "sbom_go.json"), "w") as f:
            f.write(res.stdout)
    
    # NPM SBOM (simulated for now)
    web_dir = os.path.join(REPO_ROOT, "apps/web")
    if os.path.exists(web_dir):
        res = subprocess.run("npm list --json", shell=True, capture_output=True, text=True, cwd=web_dir)
        if res.returncode == 0:
            with open(os.path.join(target_dir, "sbom_npm.json"), "w") as f:
                f.write(res.stdout)

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

if __name__ == "__main__":
    main()
