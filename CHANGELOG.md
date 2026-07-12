# Changelog

## [0.69.0](https://github.com/ERroooooR/gate/compare/v0.68.26...v0.69.0) (2026-07-12)


### Features

* add comprehensive CLI version support (fixes [#282](https://github.com/ERroooooR/gate/issues/282)) ([#561](https://github.com/ERroooooR/gate/issues/561)) ([34f103c](https://github.com/ERroooooR/gate/commit/34f103c45e2e1666d0745b421f5eecfef04aacda))
* add config subcommand to generate default config files ([#608](https://github.com/ERroooooR/gate/issues/608)) ([4ace975](https://github.com/ERroooooR/gate/commit/4ace975ae30bbbedaa88c79bea7fd29556ead5c5))
* Add DialogClear and DialogShow packets to packet tests ([2c7e2af](https://github.com/ERroooooR/gate/commit/2c7e2af97a442e4bb8482af88e724cc8445066bd))
* Add EntityArgumentType for minecraft:entity commands (PR [#78](https://github.com/ERroooooR/gate/issues/78)) ([d7779f6](https://github.com/ERroooooR/gate/commit/d7779f6c02b00f68aa90c6075c76a1ee0cdf4950))
* Add ForcedHosts configuration to config.yml and update documentation ([d98e8b3](https://github.com/ERroooooR/gate/commit/d98e8b3a7789218e143a6ffdce8d4d1ab86a13dd))
* add Minecraft 26.1 (protocol 775) support ([#679](https://github.com/ERroooooR/gate/issues/679)) ([d12d659](https://github.com/ERroooooR/gate/commit/d12d659b43fca433adc02b96399024b7dfc89ad2))
* add option to disable auto config reload ([#607](https://github.com/ERroooooR/gate/issues/607)) ([fb7d2a8](https://github.com/ERroooooR/gate/commit/fb7d2a848e294cb0504dddf8ea1723b5072f9002)), closes [#602](https://github.com/ERroooooR/gate/issues/602)
* add raknetify support ([ba048bc](https://github.com/ERroooooR/gate/commit/ba048bca04cc7907b29098aec68f9a775089da88))
* add raw raknetify passthrough ([4651404](https://github.com/ERroooooR/gate/commit/46514048c5f6d79f5cf7edfdf7d1495c865f81bd))
* add semantic version tags to Docker images ([#617](https://github.com/ERroooooR/gate/issues/617)) ([31c79d8](https://github.com/ERroooooR/gate/commit/31c79d80839d77a23709af90e1e134b54853349a)), closes [#616](https://github.com/ERroooooR/gate/issues/616)
* Add server registration and unregistration events ([#593](https://github.com/ERroooooR/gate/issues/593)) ([0ec2ed1](https://github.com/ERroooooR/gate/commit/0ec2ed1fd60e297fb686efd5436e15f6e733fc4b))
* add strategy support to lite mode ([#538](https://github.com/ERroooooR/gate/issues/538)) ([c312dd0](https://github.com/ERroooooR/gate/commit/c312dd00acb0e49884439c195dbde3eb08a83a1e))
* Add support for changing the default logger ([#536](https://github.com/ERroooooR/gate/issues/536)) ([24014d8](https://github.com/ERroooooR/gate/commit/24014d8022d8362c577fb4dbc2ee0c6c8abe359b))
* Add support for Minecraft 1.21.7 (Protocol 772) ([#546](https://github.com/ERroooooR/gate/issues/546)) ([074d3e9](https://github.com/ERroooooR/gate/commit/074d3e966d4b4b08fadfbbfe4095630f717396a9))
* Add support for offline mode players in Connect configuration ([de6f865](https://github.com/ERroooooR/gate/commit/de6f865ffbfb6a1dc4960d08bfeb4c2417801a35))
* add TCPBrutal support ([88fba1e](https://github.com/ERroooooR/gate/commit/88fba1ef12198831e1fad6281792726b7b36c49d))
* **bedrock:** default managed mode to geyserlite ([c82a7de](https://github.com/ERroooooR/gate/commit/c82a7de629ecbe94e349e2a41c07921045f5968e))
* **bedrock:** default managed mode to geyserlite ([205bb49](https://github.com/ERroooooR/gate/commit/205bb494b4bb6df7b4b05c520b2f34e7075ab6dd))
* Cookies API ([#535](https://github.com/ERroooooR/gate/issues/535)) ([2a8e179](https://github.com/ERroooooR/gate/commit/2a8e1796b5e62b2e1dacf5f09d1bb76a59c6a06f))
* **docker:** add multi-target images and bump distroless base to Debian 12 ([#568](https://github.com/ERroooooR/gate/issues/568)) ([8d91027](https://github.com/ERroooooR/gate/commit/8d910278bcebcbf8a9b568e5d1cdf1f6f75b9162))
* **docs:** add OpenAPI documentation link and update API definition formatting ([d3c6a57](https://github.com/ERroooooR/gate/commit/d3c6a57d9998b0dc73fc3251ec5f1f5ad4b14530))
* **docs:** clarify Bedrock Edition support ([#588](https://github.com/ERroooooR/gate/issues/588)) ([7cd126b](https://github.com/ERroooooR/gate/commit/7cd126b5f28721a7b18f14fc04fb2228f25f371d))
* **docs:** Use GitHub App over PAT and use octokit ([#599](https://github.com/ERroooooR/gate/issues/599)) ([959dc1b](https://github.com/ERroooooR/gate/commit/959dc1b3708bdca2e31bbdfd01c3c04df5d2e55f))
* Enhance installation script for cross-platform support ([b91fdfd](https://github.com/ERroooooR/gate/commit/b91fdfd86732999c9c95ec98c084b03bbb0ca061))
* Enhance LandingAfter component with tabbed configuration and Git clone examples ([e43bae3](https://github.com/ERroooooR/gate/commit/e43bae34dd8a1eea658ab967509b552f727f5e5b))
* harden login plugin-message and keep-alive handling ([#689](https://github.com/ERroooooR/gate/issues/689)) ([f818164](https://github.com/ERroooooR/gate/commit/f818164793cbc85c715f8f04718f10223d556db5))
* implement proper ForcedHosts functionality (fixes [#226](https://github.com/ERroooooR/gate/issues/226)) ([#560](https://github.com/ERroooooR/gate/issues/560)) ([a07c4c4](https://github.com/ERroooooR/gate/commit/a07c4c4d0f2f9ff8fc71bba92b890bfee49dee7a))
* Initial Bedrock Geyser Support ([#552](https://github.com/ERroooooR/gate/issues/552)) ([75ec7d3](https://github.com/ERroooooR/gate/commit/75ec7d3603671f53ca17c1ae2eba31fa770ad197))
* **lite:** Add support for hostname parameter routing in backend addresses ([#603](https://github.com/ERroooooR/gate/issues/603)) ([11882b0](https://github.com/ERroooooR/gate/commit/11882b0b429fab24c5ac360db0df136ed612594f))
* make ServerConnection accessible in ServerLoginPluginMessageEvent ([#664](https://github.com/ERroooooR/gate/issues/664)) ([64bd57e](https://github.com/ERroooooR/gate/commit/64bd57e25b7fcbf6959504ffcc313e047aeaf843)), closes [#663](https://github.com/ERroooooR/gate/issues/663)
* Minecraft 1.21.8 ([6faa101](https://github.com/ERroooooR/gate/commit/6faa101a11b6519f270c8bdecea9aecb822ecb5f))
* Minecraft version 1.21.6 ([#543](https://github.com/ERroooooR/gate/issues/543)) ([4cfc98e](https://github.com/ERroooooR/gate/commit/4cfc98ebd11e8347c878eab43dfe938f619d8356))
* Modern Forge (FML2/FML3) login relay for 1.13-1.20.1 ([#680](https://github.com/ERroooooR/gate/issues/680)) ([df65927](https://github.com/ERroooooR/gate/commit/df65927aa87ace5e0334130197a4dc75cfd0a09c))
* per-connection serverbound packet rate limiter ([#691](https://github.com/ERroooooR/gate/issues/691)) ([c8ed9bb](https://github.com/ERroooooR/gate/commit/c8ed9bba63b9e89091e1d5b599e3544a59dc3218))
* route dynamic servers through managed Via ([d22711d](https://github.com/ERroooooR/gate/commit/d22711dc53bb024bafa08b741f29454d332581c2))
* Support Minecraft 1.21.5 ([7eee5e6](https://github.com/ERroooooR/gate/commit/7eee5e61bfc6b4c8aaadec28981274a6ce977ceb))
* Support Minecraft Java 1.21.11 ([b3119ff](https://github.com/ERroooooR/gate/commit/b3119ffe31a5542dc351ba8de0a7f4c97dd04a3a))
* Support overriding forwarding secrets with environment variables ([#575](https://github.com/ERroooooR/gate/issues/575)) ([080575f](https://github.com/ERroooooR/gate/commit/080575f361a1c7a90f64f556e0c64df057b75077))
* Update for Minecraft 1.21.9 and 1.21.10 ([#580](https://github.com/ERroooooR/gate/issues/580)) ([40a741e](https://github.com/ERroooooR/gate/commit/40a741e6b076b6fb4c64c87c8a9357541a439677))
* use vialite auto latest subprocess runtime ([c174b31](https://github.com/ERroooooR/gate/commit/c174b3153769c33873b25a27f2e8695b977f3cb6))
* **WIP:** add opentelemetry ([#439](https://github.com/ERroooooR/gate/issues/439)) ([18eda21](https://github.com/ERroooooR/gate/commit/18eda21c439aa0d96ae3dc115a229690070999fc))


### Bug Fixes

* add musl linux installer assets ([e22541c](https://github.com/ERroooooR/gate/commit/e22541cbf57fa9841eb5ddaf19a913abe6e92c26))
* add raw raknetify write backpressure metrics ([1dcb903](https://github.com/ERroooooR/gate/commit/1dcb9030c6df8df6b9754116e4d1ba4f55e09ca2))
* adopt modern Common text components ([4f2debc](https://github.com/ERroooooR/gate/commit/4f2debcde3c083df814c078b97a590e22edcd287))
* align gate raknet with netty-raknet latest retry & timing changes ([25dae1d](https://github.com/ERroooooR/gate/commit/25dae1d33c854dc00e6944f6777ad4116578cd72))
* align gate raknet with raknetify/netty-raknet latest changes ([28aee3c](https://github.com/ERroooooR/gate/commit/28aee3cf8fda5624e4c86f400d5995f4b9b1e79c))
* avoid queued datagram pool allocation ([b29d7a6](https://github.com/ERroooooR/gate/commit/b29d7a6139e83d5a2d47e1373d272584972f6cd2))
* **bedrock:** avoid forwarding mobile ping to backend keepalive ([#771](https://github.com/ERroooooR/gate/issues/771)) ([0b6c3ba](https://github.com/ERroooooR/gate/commit/0b6c3ba369db8b01adf1a092848972122c73713f))
* **bedrock:** document engine-specific managed config ([7c21844](https://github.com/ERroooooR/gate/commit/7c218447a96dfb4f5104b382227a01293093b1d2))
* **bedrock:** document engine-specific managed config ([4d6038d](https://github.com/ERroooooR/gate/commit/4d6038d3a3590b651e8f9f662de756edb770f0af))
* **bedrock:** forward hostnames in managed java geyser ([a71a7d2](https://github.com/ERroooooR/gate/commit/a71a7d26c00cd86199ad11a89f634f9202149e5c))
* **bedrock:** forward hostnames in managed Java Geyser ([392ce14](https://github.com/ERroooooR/gate/commit/392ce14d8c3b1582b0b7226544276c47a148c252))
* **bedrock:** handle geyserlite ci platform gaps ([48c4857](https://github.com/ERroooooR/gate/commit/48c48570adf3a54fe36cce1689b362d99a936d19))
* **bedrock:** trust linked Geyser accounts safely ([c307797](https://github.com/ERroooooR/gate/commit/c307797319aa578a78efaa12ff79aa0cc6d39919))
* **bedrock:** use geyserlite windows build fix ([f266386](https://github.com/ERroooooR/gate/commit/f266386837184b433bf3ba9922bbd85f0ff0ac74))
* **bedrock:** wait for managed geyserlite readiness ([cf0a4ef](https://github.com/ERroooooR/gate/commit/cf0a4efaaee747c45381748f3160fb31a875d785))
* **bedrock:** wait for managed geyserlite readiness ([5dcbdf3](https://github.com/ERroooooR/gate/commit/5dcbdf388f73264f271f2d5f7a07b058d175bf11))
* broken .gitignore ([#596](https://github.com/ERroooooR/gate/issues/596)) ([606126c](https://github.com/ERroooooR/gate/commit/606126cedc72eac84dd6cf34fd295dcf350baaa8))
* cleanups and extra tests for SNBT/CESU-8 ([#665](https://github.com/ERroooooR/gate/issues/665)) ([#666](https://github.com/ERroooooR/gate/issues/666)) ([ca468b1](https://github.com/ERroooooR/gate/commit/ca468b1de2d62929cf6e7e185fdd274dfd6d81ec))
* close lite pipe when either side ends ([a0fbd7e](https://github.com/ERroooooR/gate/commit/a0fbd7e2f83230482b9a7520081bfd841211b04c))
* code review fixes for raknetify data races and detach handoff ([f67d229](https://github.com/ERroooooR/gate/commit/f67d2295ac5a062db8dd25749b291de52fadbd8f))
* concurrent map access, monotonic clock, defaultPingInterval correctness ([97d4760](https://github.com/ERroooooR/gate/commit/97d476030835a336a0f1887310e2e3fdc8bd281d))
* correct Play ClientBound plugin.Message packet ID for 1.21.9+ (fixes [#612](https://github.com/ERroooooR/gate/issues/612)) ([#674](https://github.com/ERroooooR/gate/issues/674)) ([f3e7f43](https://github.com/ERroooooR/gate/commit/f3e7f434818ab8f25ad5d05ab4b809fda7b37cbb))
* correct PluginMessageEvent source/target in initial connect handler ([#693](https://github.com/ERroooooR/gate/issues/693)) ([a634425](https://github.com/ERroooooR/gate/commit/a6344253ce4c86d8cb5cd77c8689058c3517b4b4))
* Correct ServerLinks packet decoding for 1.21.2+ ([074d3e9](https://github.com/ERroooooR/gate/commit/074d3e966d4b4b08fadfbbfe4095630f717396a9))
* data race in loginInboundConn during Forge login relay ([#690](https://github.com/ERroooooR/gate/issues/690)) ([1b03ac8](https://github.com/ERroooooR/gate/commit/1b03ac8226529fb201eb942bfe3bacb22978058e))
* dead detection should consider lastActivity alongside lastPongAt ([ed92604](https://github.com/ERroooooR/gate/commit/ed92604922db65579022618a3a8ede10283b5a0c))
* **deps:** automate geyserlite updates ([e934efd](https://github.com/ERroooooR/gate/commit/e934efd644bca4ede35ab849f019f3e4b4f12e96))
* **deps:** automate geyserlite updates ([6b2820b](https://github.com/ERroooooR/gate/commit/6b2820ba99c8b2ece26e9d65c8fbde5931ea376e))
* **deps:** bump geyserlite to v0.3.4 ([b574809](https://github.com/ERroooooR/gate/commit/b574809008c25776703fe63327259a625ec198ed))
* **deps:** resolve merge conflicts and update dependencies in go.mod and go.sum ([ac075d2](https://github.com/ERroooooR/gate/commit/ac075d2d3f0a8bbdfd017dd3efa11487ead647b8))
* **deps:** update dependency @bufbuild/protobuf to v1.10.1 ([#516](https://github.com/ERroooooR/gate/issues/516)) ([1f1058c](https://github.com/ERroooooR/gate/commit/1f1058c4ac72d39354c38cec24823faa5e224769))
* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.72.0.1.20241118150055.50fffb007499 ([#520](https://github.com/ERroooooR/gate/issues/520)) ([99e53a4](https://github.com/ERroooooR/gate/commit/99e53a45d1057d2bcc0e9d027279ab42c48f52f5))
* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.78.0.1.20241118150055.50fffb007499 ([#641](https://github.com/ERroooooR/gate/issues/641)) ([c4b3250](https://github.com/ERroooooR/gate/commit/c4b3250f9da279a984285cef73bde85da46af60b))
* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.81.0.1.20241118150055.50fffb007499 ([#736](https://github.com/ERroooooR/gate/issues/736)) ([3f260e5](https://github.com/ERroooooR/gate/commit/3f260e51a557b1e4fdb5369cdf163ca5c46db592))
* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.82.0.1.20241118150055.50fffb007499 ([#752](https://github.com/ERroooooR/gate/issues/752)) ([f79e95b](https://github.com/ERroooooR/gate/commit/f79e95b8d6aa5c871f1950e37eabd93d3af99404))
* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.82.1.1.20241118150055.50fffb007499 ([#832](https://github.com/ERroooooR/gate/issues/832)) ([48a2c54](https://github.com/ERroooooR/gate/commit/48a2c5434adee9c8bfef94d12eaf992d57827e62))
* **deps:** update dependency build.buf.gen:minekube_gate_protocolbuffers_java to v29.5.0.1.20241118150055.50fffb007499 ([#643](https://github.com/ERroooooR/gate/issues/643)) ([cf779b3](https://github.com/ERroooooR/gate/commit/cf779b3be9676b461d8525551cf5c6682dbb44eb))
* **deps:** update dependency build.buf.gen:minekube_gate_protocolbuffers_java to v29.5.0.2.20241118150055.50fffb007499 ([#727](https://github.com/ERroooooR/gate/issues/727)) ([e85b87c](https://github.com/ERroooooR/gate/commit/e85b87c244b12a48550f06db43b813c56fa2d4fb))
* **deps:** update dependency com.google.protobuf:protobuf-java to v4.30.2 ([#521](https://github.com/ERroooooR/gate/issues/521)) ([202a586](https://github.com/ERroooooR/gate/commit/202a58680e9352f7f571700903e0ee6a0768a390))
* **deps:** update dependency com.google.protobuf:protobuf-java to v4.33.4 ([#645](https://github.com/ERroooooR/gate/issues/645)) ([0fe4786](https://github.com/ERroooooR/gate/commit/0fe47867ac87758c51e066dd059b212a31d7eeec))
* **deps:** update dependency com.google.protobuf:protobuf-java to v4.35.1 ([#842](https://github.com/ERroooooR/gate/issues/842)) ([a236e18](https://github.com/ERroooooR/gate/commit/a236e18e69e6144717c93b9bf1eca82a10c62ec0))
* **deps:** update dependency io.grpc:grpc-kotlin-stub to v1.4.3 ([#518](https://github.com/ERroooooR/gate/issues/518)) ([de2b8b3](https://github.com/ERroooooR/gate/commit/de2b8b34caecc0df7d306330eb32af3a549a4ef8))
* **deps:** update dependency io.grpc:grpc-kotlin-stub to v1.5.0 ([#646](https://github.com/ERroooooR/gate/issues/646)) ([64e974b](https://github.com/ERroooooR/gate/commit/64e974b3fc28d241f392dcc6b0603b45a1ddb95a))
* **deps:** update dependency io.grpc:grpc-netty-shaded to v1.75.0 [security] ([#576](https://github.com/ERroooooR/gate/issues/576)) ([79f6984](https://github.com/ERroooooR/gate/commit/79f6984a910bce38a05329711fb496a0427efe06))
* **deps:** update dependency org.jetbrains.kotlinx:kotlinx-coroutines-core to v1.10.2 ([#512](https://github.com/ERroooooR/gate/issues/512)) ([24b5fcf](https://github.com/ERroooooR/gate/commit/24b5fcfe8f57ae162f006963d529e05f1ec58302))
* **deps:** Update geyserlite ([#802](https://github.com/ERroooooR/gate/issues/802)) ([eebb580](https://github.com/ERroooooR/gate/commit/eebb580bf88c79a3fe65fe6551e9c00a879b8871))
* **deps:** update geyserlite to v0.3.0 ([#722](https://github.com/ERroooooR/gate/issues/722)) ([69bfeb7](https://github.com/ERroooooR/gate/commit/69bfeb7bd3504231a547df13aebbd2f86ad040fe))
* **deps:** Update geyserlite to v0.3.10 ([#791](https://github.com/ERroooooR/gate/issues/791)) ([53bbf8f](https://github.com/ERroooooR/gate/commit/53bbf8fb37d6a5b841cbbe0d8a7acc4303369dc4))
* **deps:** Update geyserlite to v0.3.12 ([#798](https://github.com/ERroooooR/gate/issues/798)) ([f87a81e](https://github.com/ERroooooR/gate/commit/f87a81e9972862a09f3d700f9a60f96b7bc74c06))
* **deps:** update geyserlite to v0.3.14 ([a3e7ed7](https://github.com/ERroooooR/gate/commit/a3e7ed723db9b0037690d33854aa602a21255f97))
* **deps:** update geyserlite to v0.3.14 ([d854e39](https://github.com/ERroooooR/gate/commit/d854e39f4200a1d07c0b79da238c61daa1a512e8))
* **deps:** Update geyserlite to v0.3.15 ([#813](https://github.com/ERroooooR/gate/issues/813)) ([36fc507](https://github.com/ERroooooR/gate/commit/36fc507c40aad62ae921c85e9ebbff32f5669a0c))
* **deps:** update geyserlite to v0.3.16 ([#850](https://github.com/ERroooooR/gate/issues/850)) ([5803dcf](https://github.com/ERroooooR/gate/commit/5803dcfa2c433f7bec1d8c8e167783f313ce5050))
* **deps:** Update geyserlite to v0.3.17 ([#874](https://github.com/ERroooooR/gate/issues/874)) ([1494394](https://github.com/ERroooooR/gate/commit/14943941a7c4f26d823f41fb23e62877ce6b88f9))
* **deps:** Update geyserlite to v0.3.5 ([#749](https://github.com/ERroooooR/gate/issues/749)) ([e620e0b](https://github.com/ERroooooR/gate/commit/e620e0b8af0a58f788ef41c2016e3792d2fd2608))
* **deps:** update golang.org/x/exp digest to 716be56 ([#590](https://github.com/ERroooooR/gate/issues/590)) ([fbc4ad8](https://github.com/ERroooooR/gate/commit/fbc4ad89f777ede8f236901713ce33887f0225ba))
* **deps:** update golang.org/x/exp digest to c48552f ([#695](https://github.com/ERroooooR/gate/issues/695)) ([e04e3da](https://github.com/ERroooooR/gate/commit/e04e3da44608a814cead64be42bee6b555517b5a))
* **deps:** update grpc-java monorepo to v1.72.0 ([#523](https://github.com/ERroooooR/gate/issues/523)) ([b900085](https://github.com/ERroooooR/gate/commit/b900085f0397839840873c5b05423f546e03399e))
* **deps:** update grpc-java monorepo to v1.78.0 ([#649](https://github.com/ERroooooR/gate/issues/649)) ([1f05e43](https://github.com/ERroooooR/gate/commit/1f05e4315617f6fec4f08ab54acaa8db8e627eaf))
* **deps:** update grpc-java monorepo to v1.81.0 ([#738](https://github.com/ERroooooR/gate/issues/738)) ([f122ad7](https://github.com/ERroooooR/gate/commit/f122ad75a93f01783e99395005ba48ad658cb4ed))
* **deps:** update grpc-java monorepo to v1.82.0 ([#755](https://github.com/ERroooooR/gate/issues/755)) ([2feb606](https://github.com/ERroooooR/gate/commit/2feb60623c16ac7fc71046c7255544c66e83815e))
* **deps:** update grpc-java monorepo to v1.82.1 ([#836](https://github.com/ERroooooR/gate/issues/836)) ([f987505](https://github.com/ERroooooR/gate/commit/f987505c4e0bb77e72ca349f38c71631cc9fdecb))
* **deps:** update kotlinx-coroutines monorepo to v1.11.0 ([#759](https://github.com/ERroooooR/gate/issues/759)) ([6690cf1](https://github.com/ERroooooR/gate/commit/6690cf1ba8f6c0654c57ff5afced151c8b8e19c8))
* **deps:** update module buf.build/gen/go/minekube/gate/connectrpc/go to v1.19.1-20250516132630-2a0c7768e191.2 ([#652](https://github.com/ERroooooR/gate/issues/652)) ([b48faa3](https://github.com/ERroooooR/gate/commit/b48faa36d41b5304bf8a4a020811d9938073cb40))
* **deps:** update module buf.build/gen/go/minekube/gate/connectrpc/go to v1.20.0-20250516132630-2a0c7768e191.1 ([#761](https://github.com/ERroooooR/gate/issues/761)) ([0c6591e](https://github.com/ERroooooR/gate/commit/0c6591eae358c647fc0671c8cfd8e043a87d5087))
* **deps:** update module buf.build/gen/go/minekube/gate/protocolbuffers/go to v1.36.11-20250516132630-2a0c7768e191.1 ([#623](https://github.com/ERroooooR/gate/issues/623)) ([7571bb4](https://github.com/ERroooooR/gate/commit/7571bb4401e157ead7a410b99bd568b30e35a403))
* **deps:** update module buf.build/gen/go/minekube/gate/protocolbuffers/go to v1.36.6-20241120101512-f1a10b5029ce.1 ([#519](https://github.com/ERroooooR/gate/issues/519)) ([9c31fbf](https://github.com/ERroooooR/gate/commit/9c31fbf0bd96d72f727220178d763cf6cc13046e))
* **deps:** update module connectrpc.com/connect to v1.20.0 ([#763](https://github.com/ERroooooR/gate/issues/763)) ([c6741c7](https://github.com/ERroooooR/gate/commit/c6741c74fb3a2b9dfe749812cf9fbc0c4a4187f3))
* **deps:** update module connectrpc.com/otelconnect to v0.9.0 ([#653](https://github.com/ERroooooR/gate/issues/653)) ([fb8d252](https://github.com/ERroooooR/gate/commit/fb8d252a2b8133d4bfec6772a3634600fbb59866))
* **deps:** update module github.com/coder/websocket to v1.8.15 ([#766](https://github.com/ERroooooR/gate/issues/766)) ([f321f8c](https://github.com/ERroooooR/gate/commit/f321f8ca7b7ce7d60f1be40476babf3bdf756c8a))
* **deps:** update module github.com/gammazero/deque to v1.2.0 ([#654](https://github.com/ERroooooR/gate/issues/654)) ([5af988d](https://github.com/ERroooooR/gate/commit/5af988db88b574fc868a510a108767c251b84755))
* **deps:** update module github.com/gammazero/deque to v1.2.1 ([#729](https://github.com/ERroooooR/gate/issues/729)) ([1520d56](https://github.com/ERroooooR/gate/commit/1520d5612ac390191e91eaa522b02e9b5ab9467b))
* **deps:** update module github.com/go-faker/faker/v4 to v4.6.0 ([#524](https://github.com/ERroooooR/gate/issues/524)) ([9de9723](https://github.com/ERroooooR/gate/commit/9de97233a8d821a8aa8928e85c36cdbb55b2030c))
* **deps:** update module github.com/go-faker/faker/v4 to v4.6.1 ([#528](https://github.com/ERroooooR/gate/issues/528)) ([c31c248](https://github.com/ERroooooR/gate/commit/c31c248f390a19cc24b02e530604eb300b9246bf))
* **deps:** update module github.com/go-faker/faker/v4 to v4.7.0 ([#655](https://github.com/ERroooooR/gate/issues/655)) ([5ebed38](https://github.com/ERroooooR/gate/commit/5ebed38aff8050d5be7643bd40755e02246b0e12))
* **deps:** update module github.com/go-faker/faker/v4 to v4.8.0 ([#768](https://github.com/ERroooooR/gate/issues/768)) ([694b83f](https://github.com/ERroooooR/gate/commit/694b83f8c2af0d14e0ac7b3f1b15e00c4e9faddf))
* **deps:** update module github.com/go-faker/faker/v4 to v4.9.0 ([#853](https://github.com/ERroooooR/gate/issues/853)) ([9859b14](https://github.com/ERroooooR/gate/commit/9859b140eb9f0aca6ac28813cedfb14039ceac48))
* **deps:** update module github.com/gookit/color to v1.6.0 ([#656](https://github.com/ERroooooR/gate/issues/656)) ([94cf91d](https://github.com/ERroooooR/gate/commit/94cf91ddfe7c894bbadfa98496b715dfeeda9b4e))
* **deps:** update module github.com/gookit/color to v1.6.1 ([#731](https://github.com/ERroooooR/gate/issues/731)) ([da4e57d](https://github.com/ERroooooR/gate/commit/da4e57d343c19cd804ff567214a5072b8d33006b))
* **deps:** update module github.com/jellydator/ttlcache/v3 to v3.4.1 ([#804](https://github.com/ERroooooR/gate/issues/804)) ([3e3500e](https://github.com/ERroooooR/gate/commit/3e3500e9a46c113756d0d36881b193530c5c43a7))
* **deps:** update module github.com/knadh/koanf/providers/file to v1.2.1 ([#624](https://github.com/ERroooooR/gate/issues/624)) ([16eabe4](https://github.com/ERroooooR/gate/commit/16eabe44a7996d4d5b521f878524e96bb7c850f6))
* **deps:** update module github.com/pires/go-proxyproto to v0.12.0 ([#770](https://github.com/ERroooooR/gate/issues/770)) ([695f1e2](https://github.com/ERroooooR/gate/commit/695f1e2ce45e51a6b6ac0c721bd17c9da1d7429a))
* **deps:** update module github.com/pires/go-proxyproto to v0.13.0 ([#883](https://github.com/ERroooooR/gate/issues/883)) ([bc88136](https://github.com/ERroooooR/gate/commit/bc88136f27018b7d162cc9b869dd2d8a31235498))
* **deps:** update module github.com/pires/go-proxyproto to v0.9.1 ([#657](https://github.com/ERroooooR/gate/issues/657)) ([fad1080](https://github.com/ERroooooR/gate/commit/fad10801c6a45e87c30301822165bbb469eab370))
* **deps:** update module go.minekube.com/brigodier to v0.0.2 ([#625](https://github.com/ERroooooR/gate/issues/625)) ([3ff1b4b](https://github.com/ERroooooR/gate/commit/3ff1b4bfba156c90ca54467d0dcc4472a01add0f))
* **deps:** update module go.minekube.com/common to v0.4.0 ([#848](https://github.com/ERroooooR/gate/issues/848)) ([87d0c2b](https://github.com/ERroooooR/gate/commit/87d0c2b5a2335709c868dff8f95b8df67aa9f576))
* **deps:** update module go.minekube.com/gate to v0.48.1 ([#525](https://github.com/ERroooooR/gate/issues/525)) ([35e4a38](https://github.com/ERroooooR/gate/commit/35e4a38f02d4e1d8bdce20b91952dc37c970a506))
* **deps:** update module go.minekube.com/gate to v0.62.3 ([#658](https://github.com/ERroooooR/gate/issues/658)) ([657a719](https://github.com/ERroooooR/gate/commit/657a719274b8dbf00b81ae70f23729717f602f12))
* **deps:** update module go.minekube.com/gate to v0.66.24 ([#776](https://github.com/ERroooooR/gate/issues/776)) ([d428e12](https://github.com/ERroooooR/gate/commit/d428e120235b3342e8c41a090a28ad5e5e0f892b))
* **deps:** update module go.minekube.com/gate to v0.66.26 ([#782](https://github.com/ERroooooR/gate/issues/782)) ([c7745a3](https://github.com/ERroooooR/gate/commit/c7745a3630026fd87773ab281797dc7eb994cfb3))
* **deps:** update module go.minekube.com/gate to v0.66.28 ([#788](https://github.com/ERroooooR/gate/issues/788)) ([fec6844](https://github.com/ERroooooR/gate/commit/fec6844e34b50e1e82df9e0d59554d2fe14fc9ac))
* **deps:** update module go.minekube.com/gate to v0.66.30 ([#796](https://github.com/ERroooooR/gate/issues/796)) ([88f1673](https://github.com/ERroooooR/gate/commit/88f1673086827d37f3d84417aafff5a0aabae71d))
* **deps:** update module go.minekube.com/gate to v0.66.32 ([#800](https://github.com/ERroooooR/gate/issues/800)) ([84d56d5](https://github.com/ERroooooR/gate/commit/84d56d52086ad48746988d0bc8bbb868811d301a))
* **deps:** update module go.minekube.com/gate to v0.66.36 ([#806](https://github.com/ERroooooR/gate/issues/806)) ([6549dfe](https://github.com/ERroooooR/gate/commit/6549dfe607102a8906663ae1aab46b220a0e0c2e))
* **deps:** update module go.minekube.com/gate to v0.66.38 ([#816](https://github.com/ERroooooR/gate/issues/816)) ([318d0ba](https://github.com/ERroooooR/gate/commit/318d0ba66dd7551c29d8a4f295748f74d7aa5830))
* **deps:** update module go.minekube.com/gate to v0.66.39 ([#824](https://github.com/ERroooooR/gate/issues/824)) ([a842a3b](https://github.com/ERroooooR/gate/commit/a842a3b22566e331ff113dc279cd34ca1fc456b1))
* **deps:** update module go.minekube.com/gate to v0.68.11 ([#863](https://github.com/ERroooooR/gate/issues/863)) ([3d80eb0](https://github.com/ERroooooR/gate/commit/3d80eb041626b15798412862e4636d9a2c6dfb2d))
* **deps:** update module go.minekube.com/gate to v0.68.14 ([#871](https://github.com/ERroooooR/gate/issues/871)) ([b2f3158](https://github.com/ERroooooR/gate/commit/b2f31583c9c8fed5f9eb8cdd97767001f71de397))
* **deps:** update module go.minekube.com/gate to v0.68.15 ([#877](https://github.com/ERroooooR/gate/issues/877)) ([12d85a4](https://github.com/ERroooooR/gate/commit/12d85a41837076cc5353e77353a128be9fd23370))
* **deps:** update module go.minekube.com/gate to v0.68.17 ([#881](https://github.com/ERroooooR/gate/issues/881)) ([47b4ec8](https://github.com/ERroooooR/gate/commit/47b4ec823ab6cbbc16ed94cb718d647d537942dc))
* **deps:** update module go.minekube.com/gate to v0.68.20 ([#888](https://github.com/ERroooooR/gate/issues/888)) ([e9a0ac5](https://github.com/ERroooooR/gate/commit/e9a0ac50f36e85604432a81140f5c27ec5047f1d))
* **deps:** update module go.minekube.com/gate to v0.68.22 ([#896](https://github.com/ERroooooR/gate/issues/896)) ([99e5bde](https://github.com/ERroooooR/gate/commit/99e5bde347d77b22d9198d5c01a3d8e98d4e3723))
* **deps:** update module go.minekube.com/gate to v0.68.23 ([#899](https://github.com/ERroooooR/gate/issues/899)) ([6e2db65](https://github.com/ERroooooR/gate/commit/6e2db65602851fc1c745938768821b9fe828bb78))
* **deps:** update module go.minekube.com/gate to v0.68.24 ([#902](https://github.com/ERroooooR/gate/issues/902)) ([1768bd1](https://github.com/ERroooooR/gate/commit/1768bd1a05d8ec99bb0eee0e64a7d919653378d0))
* **deps:** update module go.minekube.com/gate to v0.68.25 ([#905](https://github.com/ERroooooR/gate/issues/905)) ([9dde0ac](https://github.com/ERroooooR/gate/commit/9dde0accd1d0ee0f62288dcb0fc974fd7786492e))
* **deps:** update module go.minekube.com/gate to v0.68.9 ([#856](https://github.com/ERroooooR/gate/issues/856)) ([dc6f1aa](https://github.com/ERroooooR/gate/commit/dc6f1aa473fbde53fed562e6bdbfcbac4493b5f6))
* **deps:** update module go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp to v0.64.0 ([#659](https://github.com/ERroooooR/gate/issues/659)) ([b005987](https://github.com/ERroooooR/gate/commit/b005987e75ef8da33c5b9f2ca7567a5c008339ea))
* **deps:** update module go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp to v0.69.0 ([#861](https://github.com/ERroooooR/gate/issues/861)) ([8033c48](https://github.com/ERroooooR/gate/commit/8033c4898a2150dc03f1a02035f6dda9e0e7118b))
* **deps:** update module go.opentelemetry.io/otel to v1.41.0 [security] ([#685](https://github.com/ERroooooR/gate/issues/685)) ([e5b12e6](https://github.com/ERroooooR/gate/commit/e5b12e689e7a025f660ec6059ce47bbc59d0bd95))
* **deps:** update module go.uber.org/zap to v1.27.1 ([#626](https://github.com/ERroooooR/gate/issues/626)) ([9727a30](https://github.com/ERroooooR/gate/commit/9727a30ee717a145f82c8c3938081d9c5a7cb482))
* **deps:** update module go.uber.org/zap to v1.28.0 ([#778](https://github.com/ERroooooR/gate/issues/778)) ([a85fe51](https://github.com/ERroooooR/gate/commit/a85fe518c77da8a4f3677c6e6f895e76bd61458c))
* **deps:** update module golang.org/x/net to v0.38.0 [security] ([#513](https://github.com/ERroooooR/gate/issues/513)) ([cf50511](https://github.com/ERroooooR/gate/commit/cf505118ba70cc423e004b13c91bb25e00d6fde7))
* **deps:** update module golang.org/x/net to v0.39.0 ([#526](https://github.com/ERroooooR/gate/issues/526)) ([6c5b676](https://github.com/ERroooooR/gate/commit/6c5b6766bb5255ff325de8676fbfbc3faf25ac3a))
* **deps:** update module golang.org/x/net to v0.49.0 ([#660](https://github.com/ERroooooR/gate/issues/660)) ([5a3e0f6](https://github.com/ERroooooR/gate/commit/5a3e0f6a7a34b8d4701c4070a7021e27b88facb8))
* **deps:** update module golang.org/x/sync to v0.21.0 ([#865](https://github.com/ERroooooR/gate/issues/865)) ([273b542](https://github.com/ERroooooR/gate/commit/273b542897d3a6c2d051dc60ce47479436739581))
* **deps:** update module golang.org/x/text to v0.38.0 ([#879](https://github.com/ERroooooR/gate/issues/879)) ([089e4be](https://github.com/ERroooooR/gate/commit/089e4be75e9dd2d7c36281c29a73a222e2ca2e0f))
* **deps:** update module golang.org/x/time to v0.11.0 ([#529](https://github.com/ERroooooR/gate/issues/529)) ([4fb1cee](https://github.com/ERroooooR/gate/commit/4fb1ceedf4f775966f4a79fb60c09b7ba377410f))
* **deps:** update module google.golang.org/grpc to v1.72.0 ([#530](https://github.com/ERroooooR/gate/issues/530)) ([cc649ab](https://github.com/ERroooooR/gate/commit/cc649abb4082a3cf12bd5f7d69b611f6ea867e26))
* **deps:** update module google.golang.org/grpc to v1.78.0 ([#661](https://github.com/ERroooooR/gate/issues/661)) ([398f57e](https://github.com/ERroooooR/gate/commit/398f57e2e2fc9ee9b517c644b9cae9080a337450))
* **deps:** update module google.golang.org/grpc to v1.79.3 [security] ([#678](https://github.com/ERroooooR/gate/issues/678)) ([2217f5f](https://github.com/ERroooooR/gate/commit/2217f5f357edba3b8901d100c704dc46e4bcc14b))
* **deps:** update module google.golang.org/protobuf to v1.36.11 ([#627](https://github.com/ERroooooR/gate/issues/627)) ([7bf90b4](https://github.com/ERroooooR/gate/commit/7bf90b4fcfd874b5f1627f3c41d6d15211b20a50))
* **deps:** update vialite to v0.3.0 ([#843](https://github.com/ERroooooR/gate/issues/843)) ([40e7e00](https://github.com/ERroooooR/gate/commit/40e7e0044fca39a6b66a4aca791846d2bcb94efc))
* **docker:** support managed geyserlite on arm64 ([#717](https://github.com/ERroooooR/gate/issues/717)) ([beda893](https://github.com/ERroooooR/gate/commit/beda893a11a3d158939201746c5686de13100138))
* **docker:** use glibc base images so dynamically linked gate binary execs ([#712](https://github.com/ERroooooR/gate/issues/712)) ([9d503c3](https://github.com/ERroooooR/gate/commit/9d503c348a724711835a1fcd78e8434b77024d23))
* **docs:** include ([03ecca5](https://github.com/ERroooooR/gate/commit/03ecca5eb78d361503fed26e26eb48b9ccc8e793))
* **docs:** use correct comment syntax for VitePress colored diffs in YAML blocks ([#572](https://github.com/ERroooooR/gate/issues/572)) ([23bdc97](https://github.com/ERroooooR/gate/commit/23bdc97304a3628bc68758be1e9c7ce22bbe963a))
* handle bare IPv6 addresses in proxy protocol header (fixes [#670](https://github.com/ERroooooR/gate/issues/670)) ([#673](https://github.com/ERroooooR/gate/issues/673)) ([3db7f7b](https://github.com/ERroooooR/gate/commit/3db7f7ba5433b18006cb6c0fed38445d1f369fa8))
* harden raknetify under qos loss ([f265413](https://github.com/ERroooooR/gate/commit/f265413ebda0de7210041053a063d43dd18dc2e8))
* hasRTT gate, closeSession race, remove dead code ([a3968e3](https://github.com/ERroooooR/gate/commit/a3968e3a00051c9030a5588261416b959dbcdbe6))
* ignore ErrDecoderLeftBytes in status response decoding ([#559](https://github.com/ERroooooR/gate/issues/559)) ([8ee714f](https://github.com/ERroooooR/gate/commit/8ee714f12eff1e3d400c99ac2c93b47e20023289))
* ignore proxy protocol for raknetify passthrough ([9c49246](https://github.com/ERroooooR/gate/commit/9c492468ce8e18d6761d5882e95d39792676c4b2))
* improve raknetify lite passthrough ([3f851a5](https://github.com/ERroooooR/gate/commit/3f851a5af5a752f4b679f8cf747cd327ff62ed80))
* improve raknetify passthrough throughput ([92f6d58](https://github.com/ERroooooR/gate/commit/92f6d586bc7cb53e9a5a94ff15cb3a295ba586cc))
* improve raw raknetify passthrough resilience ([7c2c91a](https://github.com/ERroooooR/gate/commit/7c2c91a7e69000dd27eb1c0c3c1682d8c2c549ae))
* increase byte array length limit for 1.7.x clients in WriteBytes17 ([#558](https://github.com/ERroooooR/gate/issues/558)) ([0dc2f12](https://github.com/ERroooooR/gate/commit/0dc2f12d0b40c113091f43f87136ebf765ee2517)), closes [#533](https://github.com/ERroooooR/gate/issues/533)
* inline dispatch-moxy-bump to resolve broken transitive workflow ref ([bf46e66](https://github.com/ERroooooR/gate/commit/bf46e66cb0e0a90b9e236f5feee9017a706845a7))
* **java/config:** validate forcedHosts against servers ([#569](https://github.com/ERroooooR/gate/issues/569)) ([98cfd60](https://github.com/ERroooooR/gate/commit/98cfd6040ce587fc77b603c53c10280e3ea0343e))
* **java:** allow nil tab list header footer ([#710](https://github.com/ERroooooR/gate/issues/710)) ([10c0f64](https://github.com/ERroooooR/gate/commit/10c0f64af2efbb56298009db868ab6573eb17d48))
* **java:** ignore untracked modern resource pack responses ([a653b71](https://github.com/ERroooooR/gate/commit/a653b7124298b0b7c36f06efd5c3e1d7224938b0))
* **java:** ignore untracked modern resource pack responses ([95cb756](https://github.com/ERroooooR/gate/commit/95cb75688e051be9b090c354c64c2a094bd08d5a))
* **java:** normalize NBT component style booleans ([486e4a8](https://github.com/ERroooooR/gate/commit/486e4a802e26168d832e0c27e1bbd48070eb54e1))
* **java:** set packet state for registry-created packets ([6cd71ba](https://github.com/ERroooooR/gate/commit/6cd71baa76eb21a9dddc93a64dd06e37760d4939))
* **java:** set packet state for registry-created packets ([0a9433b](https://github.com/ERroooooR/gate/commit/0a9433b0b12c08a0bb7e526b15c0bca7aa73ad75))
* limit docker platforms to supported linux arches ([1400ee3](https://github.com/ERroooooR/gate/commit/1400ee355eb485ae2dec699997f56418d7dafa48))
* **lite/raw-passthrough:** align session timeouts with server-side RakNet read timeout ([7b297ed](https://github.com/ERroooooR/gate/commit/7b297ed17ac95ea0e1c4b54f139b07ae93575074))
* **lite/raw-passthrough:** remove read deadline and restore 60s idle timeout ([5f2bd91](https://github.com/ERroooooR/gate/commit/5f2bd9138c651c1cc950f95acd44d552fc91184e))
* **lite/raw-passthrough:** restore per-session goroutine+channel async write to backend ([7e5f960](https://github.com/ERroooooR/gate/commit/7e5f9608e3501726e22a70c0d9148856be77a2f5))
* **lite:** fix glob cache key mutation, use QuoteMeta, and share single regex cache ([#675](https://github.com/ERroooooR/gate/issues/675)) ([c6c670b](https://github.com/ERroooooR/gate/commit/c6c670b89d6f9febe19b897e8a8c20c8be6fc7c2))
* **network:** fix AvailableCommands packet for mods ([#534](https://github.com/ERroooooR/gate/issues/534)) ([7f683a6](https://github.com/ERroooooR/gate/commit/7f683a6a2292363a0db7a16969eb789be72ea0b3))
* nil pointer dereference in session chat LastSeenMessages handling ([1f78da9](https://github.com/ERroooooR/gate/commit/1f78da990abd94a14d83c866b1e9ab46d8dd7485))
* passthrough frame read loop race with copyFrames ([f7fc2e9](https://github.com/ERroooooR/gate/commit/f7fc2e9876548ec739d7f6c4676b6f5e13c3dd2f))
* player gets kicked when using suggestions in commands ([#584](https://github.com/ERroooooR/gate/issues/584)) ([884161a](https://github.com/ERroooooR/gate/commit/884161a8cb68c6102b30b62575396c0dfd68c5ef))
* preserve raw raknetify qos marking ([f81c567](https://github.com/ERroooooR/gate/commit/f81c567ac6d6115d4a4279bb6fecb959bc93978f))
* preserve retryCount across retransmits, fix ACK timer, fix dead detection ([041c3c0](https://github.com/ERroooooR/gate/commit/041c3c04548c41f261dee84af357a904c1a35c7d))
* prevent context cancel, fragment leak, and data race in raknetify ([4f71190](https://github.com/ERroooooR/gate/commit/4f7119088f31e5944f19004c182c31b56cf6ebcc))
* prevent map sharing between config reloads causing server persistence ([#557](https://github.com/ERroooooR/gate/issues/557)) ([ad12bd9](https://github.com/ERroooooR/gate/commit/ad12bd9f6e6efa9db41409b049cbcb4ef35ca1e5))
* PreviousServer is always nil ([#541](https://github.com/ERroooooR/gate/issues/541)) ([322916a](https://github.com/ERroooooR/gate/commit/322916ac75568ef020048dfc085a12aa6214dd42))
* **proxy:** forward backend register/unregister packets to client ([#672](https://github.com/ERroooooR/gate/issues/672)) ([1c5d6c8](https://github.com/ERroooooR/gate/commit/1c5d6c8ddc61235f70244de84be9cb4cbaf4eae6))
* queue config plugin messages during backend login ([ccc6212](https://github.com/ERroooooR/gate/commit/ccc62121c3bb1b084148e65341d8d69a3a34320c))
* **raknetify:** reduce queue bloat and restore write deadline on backend ([83ba4ae](https://github.com/ERroooooR/gate/commit/83ba4ae87164a44b0d0fefbb198d8d12fc77bda7))
* **raknet:** resolve 5 reliability and correctness issues ([d944391](https://github.com/ERroooooR/gate/commit/d944391bae1bafff4c6fbe7339271cdf1b5c8455))
* remove duplicate permissions block in ci.yml ([4ec22e8](https://github.com/ERroooooR/gate/commit/4ec22e8888c336f94e6cf6e4ddbfcc411fbaff24))
* Removed `Modded Servers Config` from sidebar in favor of `Modded Servers` ([#573](https://github.com/ERroooooR/gate/issues/573)) ([6b84a42](https://github.com/ERroooooR/gate/commit/6b84a428cbf2284106cdc1bc089359d376959b37))
* resolve architecture mismatch for multi-arch docker builds ([fd083f6](https://github.com/ERroooooR/gate/commit/fd083f6f3f5ed4af63db7246cfaa5acaf337119a))
* resolve architecture mismatch for multi-arch docker builds ([a0f7da7](https://github.com/ERroooooR/gate/commit/a0f7da718477b71efe29ba1bfa062bb6d5986c75))
* resolve lite mode backend issues and add sequential strategy default ([#565](https://github.com/ERroooooR/gate/issues/565)) ([798cb65](https://github.com/ERroooooR/gate/commit/798cb659677769b698b76be738cad1db74eeae35))
* resolve nil pointer dereference warnings in forward_test.go ([#606](https://github.com/ERroooooR/gate/issues/606)) ([d1f198b](https://github.com/ERroooooR/gate/commit/d1f198b25625f15d63e6c1c0e5e3ef562730cc11))
* restore built-in raknet retransmission defaults ([c5bd524](https://github.com/ERroooooR/gate/commit/c5bd524d067f4625ff36ebc8c8babdaee30e786e))
* retain queued backend keepalive replies ([#807](https://github.com/ERroooooR/gate/issues/807)) ([03af8de](https://github.com/ERroooooR/gate/commit/03af8dea7ef15c10baeec465354af0162f95b4ab))
* Return context on ErrDecoderLeftBytes ([#579](https://github.com/ERroooooR/gate/issues/579)) ([01d87bd](https://github.com/ERroooooR/gate/commit/01d87bdfc9953cb6d8b88bc3c7cafa37de89c436))
* save client ctx before Close in passthrough detach ([31916c6](https://github.com/ERroooooR/gate/commit/31916c6259bd09ce1b254da23c074e1ec3951229))
* skip ViaLite for matching dynamic backend protocol ([27db687](https://github.com/ERroooooR/gate/commit/27db687c010383013ea106111a32340ae7a5e4ad))
* SoundEntityPacket panic and protocol desync ([#605](https://github.com/ERroooooR/gate/issues/605)) ([4592ec4](https://github.com/ERroooooR/gate/commit/4592ec40184679df0e25bf69db5b41803a3aaa15))
* strip leading slash from legacy commands to prevent double slash forwarding ([#556](https://github.com/ERroooooR/gate/issues/556)) ([5d55592](https://github.com/ERroooooR/gate/commit/5d55592a3d148f42e3f1c7d941d978c1c67be937))
* update community stats in documentation ([6eb3e56](https://github.com/ERroooooR/gate/commit/6eb3e5633dd4bbd49366a375f710018d2a8822f8))
* update gate raknet to match netty-raknet latest (fragment validation, dead detection floor, burst connect, gap timeout, deferred ACK) ([2a2d4b8](https://github.com/ERroooooR/gate/commit/2a2d4b8f33aac72e9df97725831e486aa862e8c2))
* Update member counts in LandingAfter component ([f158f91](https://github.com/ERroooooR/gate/commit/f158f91ce11a663d430cc32da933bc5e9b9df390))
* use context.WithoutCancel for backend dial after client.Close ([27aa38c](https://github.com/ERroooooR/gate/commit/27aa38c0a825f54bda68bdb0758b07113386f83a))
* v1 route hint reuse, session migration race, enqueue copy, pacing bound ([2940149](https://github.com/ERroooooR/gate/commit/2940149f9f17b7fee5bdf5e7d6fed034eb67fe8b))
* **web:** fix bedrock quick link ([#571](https://github.com/ERroooooR/gate/issues/571)) ([6b2129e](https://github.com/ERroooooR/gate/commit/6b2129e672a85698a16e312587ab1c1a47c9af1f))
* wrap SetReader with fullReader to fix large packet read issues ([#615](https://github.com/ERroooooR/gate/issues/615)) ([4b5d548](https://github.com/ERroooooR/gate/commit/4b5d54834f5ae19cae8068186694b0bc322fcbcf))


### Performance Improvements

* **raknetify:** pool packet copies in raw passthrough to reduce GC pressure ([293d50d](https://github.com/ERroooooR/gate/commit/293d50da13260a10d6553f1ccddce57d82c62790))
* **raknetify:** refactor raw passthrough to eliminate channel bufferbloat ([791b7aa](https://github.com/ERroooooR/gate/commit/791b7aa41e5326cbfff821a9d19611f3840a32c5))
* **raknetify:** remove pacing, per-write deadline, and global write lock in raw passthrough ([3b480b2](https://github.com/ERroooooR/gate/commit/3b480b2a1bd8ca5168cd6e6c649579c3da7ca84f))

## [0.68.26](https://github.com/minekube/gate/compare/v0.68.25...v0.68.26) (2026-07-05)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.25 ([#905](https://github.com/minekube/gate/issues/905)) ([9dde0ac](https://github.com/minekube/gate/commit/9dde0accd1d0ee0f62288dcb0fc974fd7786492e))

## [0.68.25](https://github.com/minekube/gate/compare/v0.68.24...v0.68.25) (2026-07-03)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.24 ([#902](https://github.com/minekube/gate/issues/902)) ([1768bd1](https://github.com/minekube/gate/commit/1768bd1a05d8ec99bb0eee0e64a7d919653378d0))

## [0.68.24](https://github.com/minekube/gate/compare/v0.68.23...v0.68.24) (2026-07-03)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.23 ([#899](https://github.com/minekube/gate/issues/899)) ([6e2db65](https://github.com/minekube/gate/commit/6e2db65602851fc1c745938768821b9fe828bb78))

## [0.68.23](https://github.com/minekube/gate/compare/v0.68.22...v0.68.23) (2026-07-02)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.22 ([#896](https://github.com/minekube/gate/issues/896)) ([99e5bde](https://github.com/minekube/gate/commit/99e5bde347d77b22d9198d5c01a3d8e98d4e3723))

## [0.68.22](https://github.com/minekube/gate/compare/v0.68.21...v0.68.22) (2026-07-02)


### Bug Fixes

* queue config plugin messages during backend login ([ccc6212](https://github.com/minekube/gate/commit/ccc62121c3bb1b084148e65341d8d69a3a34320c))

## [0.68.21](https://github.com/minekube/gate/compare/v0.68.20...v0.68.21) (2026-07-02)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.20 ([#888](https://github.com/minekube/gate/issues/888)) ([e9a0ac5](https://github.com/minekube/gate/commit/e9a0ac50f36e85604432a81140f5c27ec5047f1d))

## [0.68.20](https://github.com/minekube/gate/compare/v0.68.19...v0.68.20) (2026-07-01)


### Bug Fixes

* **deps:** update module github.com/pires/go-proxyproto to v0.13.0 ([#883](https://github.com/minekube/gate/issues/883)) ([bc88136](https://github.com/minekube/gate/commit/bc88136f27018b7d162cc9b869dd2d8a31235498))

## [0.68.19](https://github.com/minekube/gate/compare/v0.68.18...v0.68.19) (2026-07-01)


### Bug Fixes

* **java:** ignore untracked modern resource pack responses ([a653b71](https://github.com/minekube/gate/commit/a653b7124298b0b7c36f06efd5c3e1d7224938b0))
* **java:** ignore untracked modern resource pack responses ([95cb756](https://github.com/minekube/gate/commit/95cb75688e051be9b090c354c64c2a094bd08d5a))

## [0.68.18](https://github.com/minekube/gate/compare/v0.68.17...v0.68.18) (2026-07-01)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.17 ([#881](https://github.com/minekube/gate/issues/881)) ([47b4ec8](https://github.com/minekube/gate/commit/47b4ec823ab6cbbc16ed94cb718d647d537942dc))

## [0.68.17](https://github.com/minekube/gate/compare/v0.68.16...v0.68.17) (2026-06-30)


### Bug Fixes

* **deps:** update module golang.org/x/text to v0.38.0 ([#879](https://github.com/minekube/gate/issues/879)) ([089e4be](https://github.com/minekube/gate/commit/089e4be75e9dd2d7c36281c29a73a222e2ca2e0f))

## [0.68.16](https://github.com/minekube/gate/compare/v0.68.15...v0.68.16) (2026-06-30)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.15 ([#877](https://github.com/minekube/gate/issues/877)) ([12d85a4](https://github.com/minekube/gate/commit/12d85a41837076cc5353e77353a128be9fd23370))

## [0.68.15](https://github.com/minekube/gate/compare/v0.68.14...v0.68.15) (2026-06-30)


### Bug Fixes

* **deps:** Update geyserlite to v0.3.17 ([#874](https://github.com/minekube/gate/issues/874)) ([1494394](https://github.com/minekube/gate/commit/14943941a7c4f26d823f41fb23e62877ce6b88f9))
* **deps:** update module go.minekube.com/gate to v0.68.14 ([#871](https://github.com/minekube/gate/issues/871)) ([b2f3158](https://github.com/minekube/gate/commit/b2f31583c9c8fed5f9eb8cdd97767001f71de397))

## [0.68.14](https://github.com/minekube/gate/compare/v0.68.13...v0.68.14) (2026-06-29)


### Bug Fixes

* skip ViaLite for matching dynamic backend protocol ([27db687](https://github.com/minekube/gate/commit/27db687c010383013ea106111a32340ae7a5e4ad))

## [0.68.13](https://github.com/minekube/gate/compare/v0.68.12...v0.68.13) (2026-06-29)


### Bug Fixes

* **deps:** update module golang.org/x/sync to v0.21.0 ([#865](https://github.com/minekube/gate/issues/865)) ([273b542](https://github.com/minekube/gate/commit/273b542897d3a6c2d051dc60ce47479436739581))

## [0.68.12](https://github.com/minekube/gate/compare/v0.68.11...v0.68.12) (2026-06-28)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.11 ([#863](https://github.com/minekube/gate/issues/863)) ([3d80eb0](https://github.com/minekube/gate/commit/3d80eb041626b15798412862e4636d9a2c6dfb2d))

## [0.68.11](https://github.com/minekube/gate/compare/v0.68.10...v0.68.11) (2026-06-28)


### Bug Fixes

* **deps:** update module go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp to v0.69.0 ([#861](https://github.com/minekube/gate/issues/861)) ([8033c48](https://github.com/minekube/gate/commit/8033c4898a2150dc03f1a02035f6dda9e0e7118b))

## [0.68.10](https://github.com/minekube/gate/compare/v0.68.9...v0.68.10) (2026-06-28)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.68.9 ([#856](https://github.com/minekube/gate/issues/856)) ([dc6f1aa](https://github.com/minekube/gate/commit/dc6f1aa473fbde53fed562e6bdbfcbac4493b5f6))

## [0.68.9](https://github.com/minekube/gate/compare/v0.68.8...v0.68.9) (2026-06-28)


### Bug Fixes

* **deps:** update module github.com/go-faker/faker/v4 to v4.9.0 ([#853](https://github.com/minekube/gate/issues/853)) ([9859b14](https://github.com/minekube/gate/commit/9859b140eb9f0aca6ac28813cedfb14039ceac48))

## [0.68.8](https://github.com/minekube/gate/compare/v0.68.7...v0.68.8) (2026-06-28)


### Bug Fixes

* **deps:** update geyserlite to v0.3.16 ([#850](https://github.com/minekube/gate/issues/850)) ([5803dcf](https://github.com/minekube/gate/commit/5803dcfa2c433f7bec1d8c8e167783f313ce5050))

## [0.68.7](https://github.com/minekube/gate/compare/v0.68.6...v0.68.7) (2026-06-28)


### Bug Fixes

* **deps:** update module go.minekube.com/common to v0.4.0 ([#848](https://github.com/minekube/gate/issues/848)) ([87d0c2b](https://github.com/minekube/gate/commit/87d0c2b5a2335709c868dff8f95b8df67aa9f576))

## [0.68.6](https://github.com/minekube/gate/compare/v0.68.5...v0.68.6) (2026-06-27)


### Bug Fixes

* **deps:** Update geyserlite to v0.3.15 ([#813](https://github.com/minekube/gate/issues/813)) ([36fc507](https://github.com/minekube/gate/commit/36fc507c40aad62ae921c85e9ebbff32f5669a0c))

## [0.68.5](https://github.com/minekube/gate/compare/v0.68.4...v0.68.5) (2026-06-26)


### Bug Fixes

* **deps:** update dependency com.google.protobuf:protobuf-java to v4.35.1 ([#842](https://github.com/minekube/gate/issues/842)) ([a236e18](https://github.com/minekube/gate/commit/a236e18e69e6144717c93b9bf1eca82a10c62ec0))

## [0.68.4](https://github.com/minekube/gate/compare/v0.68.3...v0.68.4) (2026-06-26)


### Bug Fixes

* **deps:** update vialite to v0.3.0 ([#843](https://github.com/minekube/gate/issues/843)) ([40e7e00](https://github.com/minekube/gate/commit/40e7e0044fca39a6b66a4aca791846d2bcb94efc))

## [0.68.3](https://github.com/minekube/gate/compare/v0.68.2...v0.68.3) (2026-06-26)


### Bug Fixes

* **deps:** update grpc-java monorepo to v1.82.1 ([#836](https://github.com/minekube/gate/issues/836)) ([f987505](https://github.com/minekube/gate/commit/f987505c4e0bb77e72ca349f38c71631cc9fdecb))

## [0.68.2](https://github.com/minekube/gate/compare/v0.68.1...v0.68.2) (2026-06-26)


### Bug Fixes

* adopt modern Common text components ([4f2debc](https://github.com/minekube/gate/commit/4f2debcde3c083df814c078b97a590e22edcd287))

## [0.68.1](https://github.com/minekube/gate/compare/v0.68.0...v0.68.1) (2026-06-26)


### Bug Fixes

* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.82.1.1.20241118150055.50fffb007499 ([#832](https://github.com/minekube/gate/issues/832)) ([48a2c54](https://github.com/minekube/gate/commit/48a2c5434adee9c8bfef94d12eaf992d57827e62))

## [0.68.0](https://github.com/minekube/gate/compare/v0.67.0...v0.68.0) (2026-06-23)


### Features

* route dynamic servers through managed Via ([d22711d](https://github.com/minekube/gate/commit/d22711dc53bb024bafa08b741f29454d332581c2))

## [0.67.0](https://github.com/minekube/gate/compare/v0.66.40...v0.67.0) (2026-06-23)


### Features

* use vialite auto latest subprocess runtime ([c174b31](https://github.com/minekube/gate/commit/c174b3153769c33873b25a27f2e8695b977f3cb6))

## [0.66.40](https://github.com/minekube/gate/compare/v0.66.39...v0.66.40) (2026-06-23)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.39 ([#824](https://github.com/minekube/gate/issues/824)) ([a842a3b](https://github.com/minekube/gate/commit/a842a3b22566e331ff113dc279cd34ca1fc456b1))

## [0.66.39](https://github.com/minekube/gate/compare/v0.66.38...v0.66.39) (2026-06-23)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.38 ([#816](https://github.com/minekube/gate/issues/816)) ([318d0ba](https://github.com/minekube/gate/commit/318d0ba66dd7551c29d8a4f295748f74d7aa5830))

## [0.66.38](https://github.com/minekube/gate/compare/v0.66.37...v0.66.38) (2026-06-22)


### Bug Fixes

* **deps:** update geyserlite to v0.3.14 ([a3e7ed7](https://github.com/minekube/gate/commit/a3e7ed723db9b0037690d33854aa602a21255f97))

## [0.66.37](https://github.com/minekube/gate/compare/v0.66.36...v0.66.37) (2026-06-22)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.36 ([#806](https://github.com/minekube/gate/issues/806)) ([6549dfe](https://github.com/minekube/gate/commit/6549dfe607102a8906663ae1aab46b220a0e0c2e))

## [0.66.36](https://github.com/minekube/gate/compare/v0.66.35...v0.66.36) (2026-06-22)


### Bug Fixes

* retain queued backend keepalive replies ([#807](https://github.com/minekube/gate/issues/807)) ([03af8de](https://github.com/minekube/gate/commit/03af8dea7ef15c10baeec465354af0162f95b4ab))

## [0.66.35](https://github.com/minekube/gate/compare/v0.66.34...v0.66.35) (2026-06-22)


### Bug Fixes

* **deps:** update module github.com/jellydator/ttlcache/v3 to v3.4.1 ([#804](https://github.com/minekube/gate/issues/804)) ([3e3500e](https://github.com/minekube/gate/commit/3e3500e9a46c113756d0d36881b193530c5c43a7))

## [0.66.34](https://github.com/minekube/gate/compare/v0.66.33...v0.66.34) (2026-06-22)


### Bug Fixes

* **deps:** Update geyserlite ([#802](https://github.com/minekube/gate/issues/802)) ([eebb580](https://github.com/minekube/gate/commit/eebb580bf88c79a3fe65fe6551e9c00a879b8871))

## [0.66.33](https://github.com/minekube/gate/compare/v0.66.32...v0.66.33) (2026-06-22)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.32 ([#800](https://github.com/minekube/gate/issues/800)) ([84d56d5](https://github.com/minekube/gate/commit/84d56d52086ad48746988d0bc8bbb868811d301a))

## [0.66.32](https://github.com/minekube/gate/compare/v0.66.31...v0.66.32) (2026-06-22)


### Bug Fixes

* **deps:** Update geyserlite to v0.3.12 ([#798](https://github.com/minekube/gate/issues/798)) ([f87a81e](https://github.com/minekube/gate/commit/f87a81e9972862a09f3d700f9a60f96b7bc74c06))

## [0.66.31](https://github.com/minekube/gate/compare/v0.66.30...v0.66.31) (2026-06-22)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.30 ([#796](https://github.com/minekube/gate/issues/796)) ([88f1673](https://github.com/minekube/gate/commit/88f1673086827d37f3d84417aafff5a0aabae71d))

## [0.66.30](https://github.com/minekube/gate/compare/v0.66.29...v0.66.30) (2026-06-21)


### Bug Fixes

* **deps:** Update geyserlite to v0.3.10 ([#791](https://github.com/minekube/gate/issues/791)) ([53bbf8f](https://github.com/minekube/gate/commit/53bbf8fb37d6a5b841cbbe0d8a7acc4303369dc4))

## [0.66.29](https://github.com/minekube/gate/compare/v0.66.28...v0.66.29) (2026-06-18)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.28 ([#788](https://github.com/minekube/gate/issues/788)) ([fec6844](https://github.com/minekube/gate/commit/fec6844e34b50e1e82df9e0d59554d2fe14fc9ac))

## [0.66.28](https://github.com/minekube/gate/compare/v0.66.27...v0.66.28) (2026-06-17)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.26 ([#782](https://github.com/minekube/gate/issues/782)) ([c7745a3](https://github.com/minekube/gate/commit/c7745a3630026fd87773ab281797dc7eb994cfb3))

## [0.66.27](https://github.com/minekube/gate/compare/v0.66.26...v0.66.27) (2026-06-17)


### Bug Fixes

* add musl linux installer assets ([e22541c](https://github.com/minekube/gate/commit/e22541cbf57fa9841eb5ddaf19a913abe6e92c26))

## [0.66.26](https://github.com/minekube/gate/compare/v0.66.25...v0.66.26) (2026-06-16)


### Bug Fixes

* **deps:** update module go.uber.org/zap to v1.28.0 ([#778](https://github.com/minekube/gate/issues/778)) ([a85fe51](https://github.com/minekube/gate/commit/a85fe518c77da8a4f3677c6e6f895e76bd61458c))

## [0.66.25](https://github.com/minekube/gate/compare/v0.66.24...v0.66.25) (2026-06-16)


### Bug Fixes

* **deps:** update module go.minekube.com/gate to v0.66.24 ([#776](https://github.com/minekube/gate/issues/776)) ([d428e12](https://github.com/minekube/gate/commit/d428e120235b3342e8c41a090a28ad5e5e0f892b))

## [0.66.24](https://github.com/minekube/gate/compare/v0.66.23...v0.66.24) (2026-06-16)


### Bug Fixes

* **deps:** update module github.com/pires/go-proxyproto to v0.12.0 ([#770](https://github.com/minekube/gate/issues/770)) ([695f1e2](https://github.com/minekube/gate/commit/695f1e2ce45e51a6b6ac0c721bd17c9da1d7429a))

## [0.66.23](https://github.com/minekube/gate/compare/v0.66.22...v0.66.23) (2026-06-16)


### Bug Fixes

* **deps:** update module connectrpc.com/connect to v1.20.0 ([#763](https://github.com/minekube/gate/issues/763)) ([c6741c7](https://github.com/minekube/gate/commit/c6741c74fb3a2b9dfe749812cf9fbc0c4a4187f3))

## [0.66.22](https://github.com/minekube/gate/compare/v0.66.21...v0.66.22) (2026-06-15)


### Bug Fixes

* **bedrock:** avoid forwarding mobile ping to backend keepalive ([#771](https://github.com/minekube/gate/issues/771)) ([0b6c3ba](https://github.com/minekube/gate/commit/0b6c3ba369db8b01adf1a092848972122c73713f))

## [0.66.21](https://github.com/minekube/gate/compare/v0.66.20...v0.66.21) (2026-06-15)


### Bug Fixes

* **deps:** update module github.com/go-faker/faker/v4 to v4.8.0 ([#768](https://github.com/minekube/gate/issues/768)) ([694b83f](https://github.com/minekube/gate/commit/694b83f8c2af0d14e0ac7b3f1b15e00c4e9faddf))

## [0.66.20](https://github.com/minekube/gate/compare/v0.66.19...v0.66.20) (2026-06-15)


### Bug Fixes

* **deps:** update module github.com/coder/websocket to v1.8.15 ([#766](https://github.com/minekube/gate/issues/766)) ([f321f8c](https://github.com/minekube/gate/commit/f321f8ca7b7ce7d60f1be40476babf3bdf756c8a))

## [0.66.19](https://github.com/minekube/gate/compare/v0.66.18...v0.66.19) (2026-06-15)


### Bug Fixes

* **deps:** update module buf.build/gen/go/minekube/gate/connectrpc/go to v1.20.0-20250516132630-2a0c7768e191.1 ([#761](https://github.com/minekube/gate/issues/761)) ([0c6591e](https://github.com/minekube/gate/commit/0c6591eae358c647fc0671c8cfd8e043a87d5087))

## [0.66.18](https://github.com/minekube/gate/compare/v0.66.17...v0.66.18) (2026-06-14)


### Bug Fixes

* **deps:** update kotlinx-coroutines monorepo to v1.11.0 ([#759](https://github.com/minekube/gate/issues/759)) ([6690cf1](https://github.com/minekube/gate/commit/6690cf1ba8f6c0654c57ff5afced151c8b8e19c8))

## [0.66.17](https://github.com/minekube/gate/compare/v0.66.16...v0.66.17) (2026-06-14)


### Bug Fixes

* **deps:** update grpc-java monorepo to v1.82.0 ([#755](https://github.com/minekube/gate/issues/755)) ([2feb606](https://github.com/minekube/gate/commit/2feb60623c16ac7fc71046c7255544c66e83815e))

## [0.66.16](https://github.com/minekube/gate/compare/v0.66.15...v0.66.16) (2026-06-14)


### Bug Fixes

* **deps:** Update geyserlite to v0.3.5 ([#749](https://github.com/minekube/gate/issues/749)) ([e620e0b](https://github.com/minekube/gate/commit/e620e0b8af0a58f788ef41c2016e3792d2fd2608))

## [0.66.15](https://github.com/minekube/gate/compare/v0.66.14...v0.66.15) (2026-06-13)


### Bug Fixes

* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.82.0.1.20241118150055.50fffb007499 ([#752](https://github.com/minekube/gate/issues/752)) ([f79e95b](https://github.com/minekube/gate/commit/f79e95b8d6aa5c871f1950e37eabd93d3af99404))

## [0.66.14](https://github.com/minekube/gate/compare/v0.66.13...v0.66.14) (2026-06-13)


### Bug Fixes

* **deps:** bump geyserlite to v0.3.4 ([b574809](https://github.com/minekube/gate/commit/b574809008c25776703fe63327259a625ec198ed))

## [0.66.13](https://github.com/minekube/gate/compare/v0.66.12...v0.66.13) (2026-06-10)


### Bug Fixes

* **deps:** update grpc-java monorepo to v1.81.0 ([#738](https://github.com/minekube/gate/issues/738)) ([f122ad7](https://github.com/minekube/gate/commit/f122ad75a93f01783e99395005ba48ad658cb4ed))

## [0.66.12](https://github.com/minekube/gate/compare/v0.66.11...v0.66.12) (2026-06-10)


### Bug Fixes

* **deps:** update dependency build.buf.gen:minekube_gate_grpc_java to v1.81.0.1.20241118150055.50fffb007499 ([#736](https://github.com/minekube/gate/issues/736)) ([3f260e5](https://github.com/minekube/gate/commit/3f260e51a557b1e4fdb5369cdf163ca5c46db592))

## [0.66.11](https://github.com/minekube/gate/compare/v0.66.10...v0.66.11) (2026-06-09)


### Bug Fixes

* **deps:** update module github.com/gookit/color to v1.6.1 ([#731](https://github.com/minekube/gate/issues/731)) ([da4e57d](https://github.com/minekube/gate/commit/da4e57d343c19cd804ff567214a5072b8d33006b))

## [0.66.10](https://github.com/minekube/gate/compare/v0.66.9...v0.66.10) (2026-06-09)


### Bug Fixes

* **deps:** update module github.com/gammazero/deque to v1.2.1 ([#729](https://github.com/minekube/gate/issues/729)) ([1520d56](https://github.com/minekube/gate/commit/1520d5612ac390191e91eaa522b02e9b5ab9467b))

## [0.66.9](https://github.com/minekube/gate/compare/v0.66.8...v0.66.9) (2026-06-08)


### Bug Fixes

* **deps:** update dependency build.buf.gen:minekube_gate_protocolbuffers_java to v29.5.0.2.20241118150055.50fffb007499 ([#727](https://github.com/minekube/gate/issues/727)) ([e85b87c](https://github.com/minekube/gate/commit/e85b87c244b12a48550f06db43b813c56fa2d4fb))

## [0.66.8](https://github.com/minekube/gate/compare/v0.66.7...v0.66.8) (2026-06-08)


### Bug Fixes

* **deps:** update geyserlite to v0.3.0 ([#722](https://github.com/minekube/gate/issues/722)) ([69bfeb7](https://github.com/minekube/gate/commit/69bfeb7bd3504231a547df13aebbd2f86ad040fe))

## [0.66.7](https://github.com/minekube/gate/compare/v0.66.6...v0.66.7) (2026-06-08)


### Bug Fixes

* **docker:** support managed geyserlite on arm64 ([#717](https://github.com/minekube/gate/issues/717)) ([beda893](https://github.com/minekube/gate/commit/beda893a11a3d158939201746c5686de13100138))

## [0.66.6](https://github.com/minekube/gate/compare/v0.66.5...v0.66.6) (2026-06-06)


### Bug Fixes

* **bedrock:** trust linked Geyser accounts safely ([c307797](https://github.com/minekube/gate/commit/c307797319aa578a78efaa12ff79aa0cc6d39919))

## [0.66.5](https://github.com/minekube/gate/compare/v0.66.4...v0.66.5) (2026-06-04)


### Bug Fixes

* **docker:** use glibc base images so dynamically linked gate binary execs ([#712](https://github.com/minekube/gate/issues/712)) ([9d503c3](https://github.com/minekube/gate/commit/9d503c348a724711835a1fcd78e8434b77024d23))

## [0.66.4](https://github.com/minekube/gate/compare/v0.66.3...v0.66.4) (2026-06-04)


### Bug Fixes

* **java:** allow nil tab list header footer ([#710](https://github.com/minekube/gate/issues/710)) ([10c0f64](https://github.com/minekube/gate/commit/10c0f64af2efbb56298009db868ab6573eb17d48))

## [0.66.3](https://github.com/minekube/gate/compare/v0.66.2...v0.66.3) (2026-06-04)


### Bug Fixes

* **java:** normalize NBT component style booleans ([486e4a8](https://github.com/minekube/gate/commit/486e4a802e26168d832e0c27e1bbd48070eb54e1))

## [0.66.2](https://github.com/minekube/gate/compare/v0.66.1...v0.66.2) (2026-06-04)


### Bug Fixes

* **bedrock:** forward hostnames in managed java geyser ([a71a7d2](https://github.com/minekube/gate/commit/a71a7d26c00cd86199ad11a89f634f9202149e5c))
* **bedrock:** forward hostnames in managed Java Geyser ([392ce14](https://github.com/minekube/gate/commit/392ce14d8c3b1582b0b7226544276c47a148c252))

## [0.66.1](https://github.com/minekube/gate/compare/v0.66.0...v0.66.1) (2026-06-04)


### Bug Fixes

* **bedrock:** wait for managed geyserlite readiness ([cf0a4ef](https://github.com/minekube/gate/commit/cf0a4efaaee747c45381748f3160fb31a875d785))
* **bedrock:** wait for managed geyserlite readiness ([5dcbdf3](https://github.com/minekube/gate/commit/5dcbdf388f73264f271f2d5f7a07b058d175bf11))

## [0.66.0](https://github.com/minekube/gate/compare/v0.65.0...v0.66.0) (2026-06-04)


### Features

* **bedrock:** default managed mode to geyserlite ([c82a7de](https://github.com/minekube/gate/commit/c82a7de629ecbe94e349e2a41c07921045f5968e))
* **bedrock:** default managed mode to geyserlite ([205bb49](https://github.com/minekube/gate/commit/205bb494b4bb6df7b4b05c520b2f34e7075ab6dd))


### Bug Fixes

* **bedrock:** document engine-specific managed config ([7c21844](https://github.com/minekube/gate/commit/7c218447a96dfb4f5104b382227a01293093b1d2))
* **bedrock:** document engine-specific managed config ([4d6038d](https://github.com/minekube/gate/commit/4d6038d3a3590b651e8f9f662de756edb770f0af))
* **bedrock:** handle geyserlite ci platform gaps ([48c4857](https://github.com/minekube/gate/commit/48c48570adf3a54fe36cce1689b362d99a936d19))
* **bedrock:** use geyserlite windows build fix ([f266386](https://github.com/minekube/gate/commit/f266386837184b433bf3ba9922bbd85f0ff0ac74))
* correct PluginMessageEvent source/target in initial connect handler ([#693](https://github.com/minekube/gate/issues/693)) ([a634425](https://github.com/minekube/gate/commit/a6344253ce4c86d8cb5cd77c8689058c3517b4b4))
* **deps:** automate geyserlite updates ([e934efd](https://github.com/minekube/gate/commit/e934efd644bca4ede35ab849f019f3e4b4f12e96))
* **deps:** automate geyserlite updates ([6b2820b](https://github.com/minekube/gate/commit/6b2820ba99c8b2ece26e9d65c8fbde5931ea376e))
* **java:** set packet state for registry-created packets ([6cd71ba](https://github.com/minekube/gate/commit/6cd71baa76eb21a9dddc93a64dd06e37760d4939))
* **java:** set packet state for registry-created packets ([0a9433b](https://github.com/minekube/gate/commit/0a9433b0b12c08a0bb7e526b15c0bca7aa73ad75))
