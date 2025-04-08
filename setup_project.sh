#!/bin/bash

# Tentukan nama direktori root proyek
PROJECT_ROOT="project"

# Buat daftar direktori yang akan dibuat
DIRS=(    
    "controller"
    "core"
    "gateway"
    "usecase"
    "middleware"
    "wiring"
)

# Buat direktori root proyek
mkdir -p "$PROJECT_ROOT"

# Buat subdirektori di dalam project/
for dir in "${DIRS[@]}"; do
    mkdir -p "$PROJECT_ROOT/$dir"
done

# Buat file main.go kosong
touch "$PROJECT_ROOT/main.go"

# Isi file main.go
echo "package main\n\nimport (\n\n)\n\nfunc main() {\n\n}" > "$PROJECT_ROOT/main.go"

# Inisialisasi go work
go work init

cd "$PROJECT_ROOT"

# Inisialisasi go mod
go mod init $PROJECT_ROOT

cd ..

go work use $PROJECT_ROOT

echo "Struktur proyek berhasil dibuat di '$PROJECT_ROOT'"
