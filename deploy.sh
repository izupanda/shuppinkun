#!/bin/bash

# リポジトリのディレクトリに移動
cd /path/to/your/repository

# 最新のコードをプル
git pull origin main

# 任意: アプリケーションを再起動 (例: Goアプリケーションの場合)
# pkill your_app_name
# go build -o your_app_name
# ./your_app_name &

# 他の必要なデプロイタスクをここに追加
