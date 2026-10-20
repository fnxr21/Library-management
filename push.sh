#!/usr/bin/env sh
# Commit staged changes and push the current branch.
# Stage files yourself first (git add <path>); this script never runs `git add .`.
set -e

branch=$(git rev-parse --abbrev-ref HEAD)
case "$branch" in
  main|master)
    echo "Refusing to push to $branch. Push a feature branch and open a pull request."
    exit 1
    ;;
esac

if git diff --cached --quiet; then
  echo 'Nothing staged. Run git add <path> first.'
  exit 1
fi

git status --short

echo 'Enter the commit message (e.g. feat(book): add stock endpoint):'
read -r commitMessage

git commit -m "$commitMessage"

git push -u origin "$branch"

echo "Pushed $branch"
