# Changelog

## [v0.1.2](https://github.com/terra-money/mantlemint/tree/v0.1.2) (2022-04-14)

This release contains a depencency upgrade for [core](https://github.com/terra-money/core)@0.5.17...@0.5.18. Apart from the dependency bump all functionalities should be exactly the same as v0.1.1.

You may need to rebuild contracts cache. Referencing from core@0.5.18 [release note](https://github.com/terra-money/core/releases/tag/v0.5.18):

> Release Note
>
> This release contains a wasmer version bump from v2.0.0 to v2.2.1. The wasm caches of these two versions are not compatible, thus rebuilding is required.
>
> To avoid possible sync delays due to the runtime rebuilding overhead, it is highly recommended that node operators rebuild their wasm cache with the [cosmwasm-cache-rebuilder](https://github.com/terra-money/cosmwasm-cache-rebuilder) before replacing terrad runtime to v0.5.18.
>
> Node upgrade instructions
>
> 1. Rebuild your wasm cache using the cosmwasm-cache-rebuilder. This rebuilder can be run simultaneously without killing a running terrad process.
> 2. You can ignore the file already open error. If other errors occur (no disk space, etc), it is safe to run the rebuilder multiple times.
> 3. When the rebuilder is finished, update terrad to v0.5.18 and restart.
> 4. The rebuilder creates a `$TERRA_HOME/data/wasm/cache/modules/v3-wasmer1` directory. The `$TERRA_HOME/data/wasm/cache/modules/v1` can be deleted after updating terrad to v0.5.18.


[Full Changelog](https://github.com/terra-money/mantlemint/compare/v0.1.1...v0.1.2)

**Merged pull requests:**

- deps\(core\): bump core@0.5.18 [\#46](https://github.com/terra-money/mantlemint/pull/46) ([kjessec](https://github.com/kjessec))

## [v0.1.1](https://github.com/terra-money/mantlemint/tree/v0.1.1) (2022-04-13)

[Full Changelog](https://github.com/terra-money/mantlemint/compare/v0.1.0...v0.1.1)

**Closed issues:**

- Mantlemint continoulsy crashes  [\#41](https://github.com/terra-money/mantlemint/issues/41)
- Potentially outdated README [\#40](https://github.com/terra-money/mantlemint/issues/40)
- Low Priority - High Memory Consumption, mantlemint get OOM killed \(32GB\) [\#38](https://github.com/terra-money/mantlemint/issues/38)
- Low Priority - Add --x-crisis-skip-assert-invariants [\#35](https://github.com/terra-money/mantlemint/issues/35)

**Merged pull requests:**

- Bump CosmWasm to 0.16.6 [\#43](https://github.com/terra-money/mantlemint/pull/43) ([hanjukim](https://github.com/hanjukim))
- docs: update crisis skip, faq [\#39](https://github.com/terra-money/mantlemint/pull/39) ([kjessec](https://github.com/kjessec))
- Upgrade tm-db to v0.6.4-performance.7 [\#37](https://github.com/terra-money/mantlemint/pull/37) ([hanjukim](https://github.com/hanjukim))
- fix\(app\): use viper as appOpts instead of Default [\#36](https://github.com/terra-money/mantlemint/pull/36) ([kjessec](https://github.com/kjessec))

## [v0.1.0](https://github.com/terra-money/mantlemint/tree/v0.1.0) (2022-03-11)

[Full Changelog](https://github.com/terra-money/mantlemint/compare/51dc2bba0eeac5bf29b7d213b53ac93846077449...v0.1.0)

**Merged pull requests:**

- \[WIP\] release/v0.1.0 [\#34](https://github.com/terra-money/mantlemint/pull/34) ([kjessec](https://github.com/kjessec))
- feat: remove readlock on query [\#32](https://github.com/terra-money/mantlemint/pull/32) ([jeffwoooo](https://github.com/jeffwoooo))
- feat: logging rollback batch metric [\#31](https://github.com/terra-money/mantlemint/pull/31) ([jeffwoooo](https://github.com/jeffwoooo))
- feat: skip cache handling for health check [\#30](https://github.com/terra-money/mantlemint/pull/30) ([jeffwoooo](https://github.com/jeffwoooo))
- feat: recover block curruption on poison error [\#28](https://github.com/terra-money/mantlemint/pull/28) ([jeffwoooo](https://github.com/jeffwoooo))
- deps: core@0.5.16 [\#27](https://github.com/terra-money/mantlemint/pull/27) ([kjessec](https://github.com/kjessec))
- feat: apply readlock on concurrent client [\#26](https://github.com/terra-money/mantlemint/pull/26) ([kjessec](https://github.com/kjessec))
- feat: restore cache mutex [\#25](https://github.com/terra-money/mantlemint/pull/25) ([jeffwoooo](https://github.com/jeffwoooo))
- feat: remove lazy sync [\#22](https://github.com/terra-money/mantlemint/pull/22) ([jeffwoooo](https://github.com/jeffwoooo))
- feat: mimalloc build in alpine build [\#21](https://github.com/terra-money/mantlemint/pull/21) ([jeffwoooo](https://github.com/jeffwoooo))
- fix: waiting channel in lock causes halt [\#17](https://github.com/terra-money/mantlemint/pull/17) ([jeffwoooo](https://github.com/jeffwoooo))
-  hotfix: crash caused by concurrent map read [\#16](https://github.com/terra-money/mantlemint/pull/16) ([jeffwoooo](https://github.com/jeffwoooo))
- fix: resurrect indexer endpoints [\#15](https://github.com/terra-money/mantlemint/pull/15) ([kjessec](https://github.com/kjessec))
- hotfix: crash caused by concurrent map read write [\#14](https://github.com/terra-money/mantlemint/pull/14) ([jeffwoooo](https://github.com/jeffwoooo))
- feat: GRPC supports query's height param [\#13](https://github.com/terra-money/mantlemint/pull/13) ([jeffwoooo](https://github.com/jeffwoooo))
- hotfix: panic on starting iterator with same start and end [\#11](https://github.com/terra-money/mantlemint/pull/11) ([jeffwoooo](https://github.com/jeffwoooo))
- chore: public release [\#10](https://github.com/terra-money/mantlemint/pull/10) ([kjessec](https://github.com/kjessec))
- \[WIP\] docs [\#7](https://github.com/terra-money/mantlemint/pull/7) ([kjessec](https://github.com/kjessec))
- feat: fork tendermint client creator to NOT use mutex [\#6](https://github.com/terra-money/mantlemint/pull/6) ([kjessec](https://github.com/kjessec))
- feat: add snappy compressor db for indexer [\#5](https://github.com/terra-money/mantlemint/pull/5) ([kjessec](https://github.com/kjessec))
- feat: height limited query with leveldb [\#3](https://github.com/terra-money/mantlemint/pull/3) ([jeffwoooo](https://github.com/jeffwoooo))



\* *This Changelog was automatically generated by [github_changelog_generator](https://github.com/github-changelog-generator/github-changelog-generator)*
