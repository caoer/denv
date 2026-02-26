# Changelog

All notable changes to denv will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Bug Fixes

- Resolve CI/CD issues - linting errors and test failures ([3b75147](https://github.com/caoer/denv/commit/3b7514703a9eaca6940f32d3efee41c1de1d4aa8))
- Resolve remaining CI/CD issues ([45e1223](https://github.com/caoer/denv/commit/45e1223128c421cf65f071d475692e6ddbed16a3))
- Resolve remaining test file linting issues ([57153cb](https://github.com/caoer/denv/commit/57153cb5a61b80d9d63ddeb51fd134056e39c3aa))
- Resolve final CI/CD issues ([19ad69a](https://github.com/caoer/denv/commit/19ad69a9edb5e2fab63ee93cc10104046e649046))
- Resolve Windows build and final linting issues ([f26e461](https://github.com/caoer/denv/commit/f26e4614291be4d97f45b10995ae7ca846f755bc))
- Final linting issues in test files ([7dac90c](https://github.com/caoer/denv/commit/7dac90c3a9dc5772d6aea7f284889ad8f16987cf))
- Resolve final test issues ([ae545e7](https://github.com/caoer/denv/commit/ae545e7e4a4e582b556fba7395f12fdf4ba441c3))
- Resolve remaining CI issues ([5c940c0](https://github.com/caoer/denv/commit/5c940c067e1b5b759698fe9ae7a254e0d8c5d586))
- Final linting issue in symlinks_test.go ([e70d245](https://github.com/caoer/denv/commit/e70d245930f5a607c639656a58dcecbf4c9c15d0))
- Resolve gosec security issues ([c009a01](https://github.com/caoer/denv/commit/c009a011da0293ffbcc30a73e56100d15eb3aff9))
- Correct hash.Write return value handling ([5309cb2](https://github.com/caoer/denv/commit/5309cb20c60c59dcaff0efa6a8493ffa238108fd))
- Suppress false positive gosec G115 warnings ([3aee9ca](https://github.com/caoer/denv/commit/3aee9caaf64aad737dde2418ef05375fe6e6f825))

### CI/CD

- Remove Windows from CI workflow matrix ([3a50418](https://github.com/caoer/denv/commit/3a504184586e6f4fd5198c00ae8e0483b478aa4e))
- Configure gosec to not fail CI on security findings ([c8a9927](https://github.com/caoer/denv/commit/c8a9927c68769f19cf6a016308c05b33e4431e65))
- Simplify gosec security scan step ([b0b5bcb](https://github.com/caoer/denv/commit/b0b5bcb3a9a8747bd0d7df59d331012181caa952))
- Update Go version to 1.24 to match go.mod ([8435216](https://github.com/caoer/denv/commit/843521677d8f3b99e9a7b660203c835178e0cd41))
- Bump the github-actions group across 1 directory with 8 updates ([7dff603](https://github.com/caoer/denv/commit/7dff603521cad197b05d6e6c7f46eb160acebc88))

### Features

- Add default ignored directories and config update command ([e9f43fc](https://github.com/caoer/denv/commit/e9f43fc9007c906d8b8c4d4a23a6a8d1f4f3d187))

### Miscellaneous

- Update changelog [skip ci] ([eef2258](https://github.com/caoer/denv/commit/eef22586752ed03047256934fa35f0c197a7671c))
- Update changelog [skip ci] ([61d41ac](https://github.com/caoer/denv/commit/61d41ace2e55a02623f1100be874a441cbbe03c4))
- Update changelog [skip ci] ([76ce0b0](https://github.com/caoer/denv/commit/76ce0b07cb3c1c29be2dd8aae5a0c2c1b32b47de))
- Update changelog [skip ci] ([b77f20c](https://github.com/caoer/denv/commit/b77f20c0a28bbd19183f54509f027a878f4cc362))
- Update changelog [skip ci] ([bc41966](https://github.com/caoer/denv/commit/bc4196699de21ebae31408135abc05550033c3ba))
- Update changelog [skip ci] ([6b167ef](https://github.com/caoer/denv/commit/6b167efae17c8072df2ee4dcb672e7d3b96325d2))
- Update changelog [skip ci] ([673726b](https://github.com/caoer/denv/commit/673726bf8744b815278a78b9fa3ac6709ccbad8d))
- Update changelog [skip ci] ([3fb8fa2](https://github.com/caoer/denv/commit/3fb8fa215ea2c7e4996ab6d3605a0848f8cf232d))
- Update changelog [skip ci] ([5639e55](https://github.com/caoer/denv/commit/5639e554e2291f30f21da501080c20e550a73403))
- Update repository references to use 'caoer' instead of 'yourusername' ([79b6acb](https://github.com/caoer/denv/commit/79b6acbaa15991e735d7a8d4d0de7374a9707120))
- Update changelog [skip ci] ([aa1b74e](https://github.com/caoer/denv/commit/aa1b74e4f18f40fe8f4a9ba53dcc7c6ccf9d575d))
- Update changelog [skip ci] ([5c496b2](https://github.com/caoer/denv/commit/5c496b21aac1f062acbf5061a79c248daeef9038))
- Update changelog [skip ci] ([7ee8ffc](https://github.com/caoer/denv/commit/7ee8ffcfe78937b16b954e88c71dcdb8a499d707))
- Update changelog [skip ci] ([dbb52ce](https://github.com/caoer/denv/commit/dbb52ce71edab78f1be8ec4bc6d91bdbd21e6ed1))
- Update changelog [skip ci] ([5665f04](https://github.com/caoer/denv/commit/5665f0452583d32b35a5427eaf812bec2a590fc9))
- Update changelog [skip ci] ([9766b46](https://github.com/caoer/denv/commit/9766b468e523aec18b755cf5dc7fd24a68ccd8a3))
- Update changelog [skip ci] ([8cdbe85](https://github.com/caoer/denv/commit/8cdbe85fc405ff8249e8297aa6c3655cbf81aa11))
- Update changelog [skip ci] ([7ac0d1d](https://github.com/caoer/denv/commit/7ac0d1da23c9b9a4083611c08d105b2b9b977909))
- Update changelog [skip ci] ([6a710af](https://github.com/caoer/denv/commit/6a710af0a6f3b2a55a6b349d77f131903a835a12))
- Update changelog [skip ci] ([b49d59f](https://github.com/caoer/denv/commit/b49d59fa8feab1ea5816ce42bf8bcd0c1097c7b5))
- Update changelog [skip ci] ([c14b629](https://github.com/caoer/denv/commit/c14b6297d0802b05315143d7440333145e8ad570))

## [1.0.0] - 2025-08-13

### Bug Fixes

- Exit with code 0 when showing help ([a2849e7](https://github.com/caoer/denv/commit/a2849e7fda3ec70ae192d18fd85144d70a2cb1af))

### Miscellaneous

- 1.0.0 [skip ci] ([fe86176](https://github.com/caoer/denv/commit/fe8617684899534dc069fa70b46975ff433a938b))

---
Generated by [git-cliff](https://github.com/orhun/git-cliff).
