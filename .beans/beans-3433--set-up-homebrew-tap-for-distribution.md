---
title: Set up Homebrew tap for distribution
status: done
created_at: 2025-12-06T16:31:34Z
updated_at: 2025-12-06T17:05:26Z
---


## Summary

Allow users to install beans via Homebrew by setting up a tap repository and configuring goreleaser to auto-publish formulas.

## Steps

### 1. Create Homebrew Tap Repository

Create a new GitHub repo named `homebrew-beans` (the `homebrew-` prefix enables the shorthand `brew install hmans/beans/beans`).

### 2. Add `brews` Section to `.goreleaser.yaml`

```yaml
brews:
  - repository:
      owner: hmans
      name: homebrew-beans
    homepage: "https://github.com/hmans/beans"
    description: "Agentic-first issue tracker"
    license: "MIT" # adjust as needed
```

### 3. Configure GitHub Token

Ensure the `GITHUB_TOKEN` used in CI has write access to the tap repository.

### 4. Test Release

Run a release and verify the formula is pushed to the tap repo.

## Resources

- [GoReleaser Homebrew Taps Documentation](https://goreleaser.com/customization/homebrew/)
- [goreleaser/homebrew-tap example](https://github.com/goreleaser/homebrew-tap)

## Checklist

- [ ] Create `hmans/homebrew-beans` repository on GitHub
- [ ] Add `brews` section to `.goreleaser.yaml`
- [ ] Verify GitHub token permissions in CI
- [ ] Create a test release
- [ ] Verify formula appears in tap repo
- [ ] Test installation with `brew install hmans/beans/beans`
