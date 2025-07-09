* Use debug version of envoy with assertions
* run the control-plane with `./scripts/build_and_run_control_plane.sh`
* run envoy with `./scripts/run_envoy.sh`
* then to update vhost config run `go run cli/*.go -config examples/dynamic_vhost_cluster1.json -action update`
* make a request to the vhost you just added `./make_request.sh 10002`
* change the vhost defintion `go run cli/*.go  -config examples/dynamic_vhost_cluster2.json -action update`

It should have the following assertion error

```
[2025-07-09 10:48:08.894][17679788][debug][config] [source/extensions/config_subscription/grpc/new_grpc_mux_impl.cc:158] Received DeltaDiscoveryResponse for type.googleapis.com/envoy.config.route.v3.VirtualHost at version 6
[2025-07-09 10:48:08.895][17679788][debug][upstream] [source/common/upstream/od_cds_api_impl.cc:39] odcds: creating subscription from config source
[2025-07-09 10:48:08.896][17679788][debug][router] [source/common/router/vhds.cc:108] vhds: loading new configuration: config_name=dynamic_route hash=15161633098654182640
[2025-07-09 10:48:08.896][17679788][debug][config] [source/extensions/config_subscription/grpc/delta_subscription_state.cc:289] Delta config for type.googleapis.com/envoy.config.route.v3.VirtualHost accepted with 1 resources added, 0 removed
[2025-07-09 10:48:08.897][17679847][critical][assert] [source/common/http/async_client_impl.cc:257] assert failure: dispatcher().isThreadSafe().
[2025-07-09 10:48:08.897][17679847][error][envoy_bug] [./source/common/common/assert.h:38] stacktrace for envoy bug
[2025-07-09 10:48:08.920][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #0 Envoy::Http::AsyncStreamImpl::sendData() [0x106ab5dc0]
[2025-07-09 10:48:08.923][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #1 Envoy::Grpc::AsyncStreamImpl::sendMessageRaw() [0x106a9e7b0]
[2025-07-09 10:48:08.925][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #2 Envoy::Grpc::Internal::sendMessageUntyped() [0x106aaaea0]
[2025-07-09 10:48:08.926][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #3 Envoy::Grpc::AsyncStream<>::sendMessage() [0x1056dcd5c]
[2025-07-09 10:48:08.928][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #4 Envoy::Config::GrpcStream<>::sendMessage() [0x1056d7824]
[2025-07-09 10:48:08.930][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #5 Envoy::Config::NewGrpcMuxImpl::trySendDiscoveryRequests() [0x1056c83c8]
[2025-07-09 10:48:08.931][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #6 Envoy::Config::NewGrpcMuxImpl::updateWatch() [0x1056cb04c]
[2025-07-09 10:48:08.933][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #7 Envoy::Config::NewGrpcMuxImpl::removeWatch() [0x1056cbd10]
[2025-07-09 10:48:08.934][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #8 Envoy::Config::NewGrpcMuxImpl::WatchImpl::remove() [0x1056eea5c]
[2025-07-09 10:48:08.936][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #9 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1056ee9f8]
[2025-07-09 10:48:08.937][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #10 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1056ee950]
[2025-07-09 10:48:08.939][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #11 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1056ee97c]
[2025-07-09 10:48:08.940][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #12 std::__1::default_delete<>::operator()[abi:ne190102]() [0x105681744]
[2025-07-09 10:48:08.941][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #13 std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1056816b8]
[2025-07-09 10:48:08.943][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #14 std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105681648]
[2025-07-09 10:48:08.944][17679847][error][envoy_bug] [./source/common/common/assert.h:43] #15 std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x10567f1dc]
[2025-07-09 10:48:08.944][17679847][critical][backtrace] [./source/server/backtrace.h:129] Caught Abort trap: 6, suspect faulting address 0x19f9bd388
[2025-07-09 10:48:08.944][17679847][critical][backtrace] [./source/server/backtrace.h:113] Backtrace (use tools/stack_decode.py to get line numbers):
[2025-07-09 10:48:08.944][17679847][critical][backtrace] [./source/server/backtrace.h:114] Envoy version: a17185b6617c6a631607ba32cda4a2d3c32f9b3b/1.34.1/Modified/DEBUG/BoringSSL
[2025-07-09 10:48:08.946][17679847][critical][backtrace] [./source/server/backtrace.h:121] #0: Envoy::SignalAction::sigHandler() [0x107f057c8]
[2025-07-09 10:48:08.946][17679847][critical][backtrace] [./source/server/backtrace.h:121] #1: _sigtramp [0x19fa30624]
[2025-07-09 10:48:08.946][17679847][critical][backtrace] [./source/server/backtrace.h:121] #2: pthread_kill [0x19f9f688c]
[2025-07-09 10:48:08.946][17679847][critical][backtrace] [./source/server/backtrace.h:121] #3: abort [0x19f8ffc60]
[2025-07-09 10:48:08.947][17679847][critical][backtrace] [./source/server/backtrace.h:121] #4: Envoy::Http::AsyncStreamImpl::sendData() [0x106ab5dd4]
[2025-07-09 10:48:08.948][17679847][critical][backtrace] [./source/server/backtrace.h:121] #5: Envoy::Grpc::AsyncStreamImpl::sendMessageRaw() [0x106a9e7b0]
[2025-07-09 10:48:08.950][17679847][critical][backtrace] [./source/server/backtrace.h:121] #6: Envoy::Grpc::Internal::sendMessageUntyped() [0x106aaaea0]
[2025-07-09 10:48:08.951][17679847][critical][backtrace] [./source/server/backtrace.h:121] #7: Envoy::Grpc::AsyncStream<>::sendMessage() [0x1056dcd5c]
[2025-07-09 10:48:08.953][17679847][critical][backtrace] [./source/server/backtrace.h:121] #8: Envoy::Config::GrpcStream<>::sendMessage() [0x1056d7824]
[2025-07-09 10:48:08.954][17679847][critical][backtrace] [./source/server/backtrace.h:121] #9: Envoy::Config::NewGrpcMuxImpl::trySendDiscoveryRequests() [0x1056c83c8]
[2025-07-09 10:48:08.955][17679847][critical][backtrace] [./source/server/backtrace.h:121] #10: Envoy::Config::NewGrpcMuxImpl::updateWatch() [0x1056cb04c]
[2025-07-09 10:48:08.957][17679847][critical][backtrace] [./source/server/backtrace.h:121] #11: Envoy::Config::NewGrpcMuxImpl::removeWatch() [0x1056cbd10]
[2025-07-09 10:48:08.958][17679847][critical][backtrace] [./source/server/backtrace.h:121] #12: Envoy::Config::NewGrpcMuxImpl::WatchImpl::remove() [0x1056eea5c]
[2025-07-09 10:48:08.959][17679847][critical][backtrace] [./source/server/backtrace.h:121] #13: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1056ee9f8]
[2025-07-09 10:48:08.961][17679847][critical][backtrace] [./source/server/backtrace.h:121] #14: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1056ee950]
[2025-07-09 10:48:08.962][17679847][critical][backtrace] [./source/server/backtrace.h:121] #15: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1056ee97c]
[2025-07-09 10:48:08.963][17679847][critical][backtrace] [./source/server/backtrace.h:121] #16: std::__1::default_delete<>::operator()[abi:ne190102]() [0x105681744]
[2025-07-09 10:48:08.965][17679847][critical][backtrace] [./source/server/backtrace.h:121] #17: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1056816b8]
[2025-07-09 10:48:08.966][17679847][critical][backtrace] [./source/server/backtrace.h:121] #18: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105681648]
[2025-07-09 10:48:08.967][17679847][critical][backtrace] [./source/server/backtrace.h:121] #19: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x10567f1dc]
[2025-07-09 10:48:08.969][17679847][critical][backtrace] [./source/server/backtrace.h:121] #20: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x105680dcc]
[2025-07-09 10:48:08.970][17679847][critical][backtrace] [./source/server/backtrace.h:121] #21: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x105681078]
[2025-07-09 10:48:08.971][17679847][critical][backtrace] [./source/server/backtrace.h:121] #22: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x1056810a4]
[2025-07-09 10:48:08.972][17679847][critical][backtrace] [./source/server/backtrace.h:121] #23: std::__1::default_delete<>::operator()[abi:ne190102]() [0x102444a64]
[2025-07-09 10:48:08.973][17679847][critical][backtrace] [./source/server/backtrace.h:121] #24: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1024449d8]
[2025-07-09 10:48:08.975][17679847][critical][backtrace] [./source/server/backtrace.h:121] #25: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x102444968]
[2025-07-09 10:48:08.976][17679847][critical][backtrace] [./source/server/backtrace.h:121] #26: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x102428094]
[2025-07-09 10:48:08.977][17679847][critical][backtrace] [./source/server/backtrace.h:121] #27: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x103c0a330]
[2025-07-09 10:48:08.978][17679847][critical][backtrace] [./source/server/backtrace.h:121] #28: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x103c09c44]
[2025-07-09 10:48:08.979][17679847][critical][backtrace] [./source/server/backtrace.h:121] #29: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x103c09c70]
[2025-07-09 10:48:08.981][17679847][critical][backtrace] [./source/server/backtrace.h:121] #30: std::__1::default_delete<>::operator()[abi:ne190102]() [0x103c0aad4]
[2025-07-09 10:48:08.982][17679847][critical][backtrace] [./source/server/backtrace.h:121] #31: std::__1::__shared_ptr_pointer<>::__on_zero_shared() [0x103c0a86c]
[2025-07-09 10:48:08.983][17679847][critical][backtrace] [./source/server/backtrace.h:121] #32: std::__1::__shared_count::__release_shared[abi:ne190102]() [0x1022b2940]
[2025-07-09 10:48:08.984][17679847][critical][backtrace] [./source/server/backtrace.h:121] #33: std::__1::__shared_weak_count::__release_shared[abi:ne190102]() [0x1022b28e4]
[2025-07-09 10:48:08.985][17679847][critical][backtrace] [./source/server/backtrace.h:121] #34: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x103c0adec]
[2025-07-09 10:48:08.986][17679847][critical][backtrace] [./source/server/backtrace.h:121] #35: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x103c07300]
[2025-07-09 10:48:08.988][17679847][critical][backtrace] [./source/server/backtrace.h:121] #36: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x105bf6528]
[2025-07-09 10:48:08.989][17679847][critical][backtrace] [./source/server/backtrace.h:121] #37: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x105bf6358]
[2025-07-09 10:48:08.990][17679847][critical][backtrace] [./source/server/backtrace.h:121] #38: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x105bf6384]
[2025-07-09 10:48:08.991][17679847][critical][backtrace] [./source/server/backtrace.h:121] #39: std::__1::default_delete<>::operator()[abi:ne190102]() [0x1033e2920]
[2025-07-09 10:48:08.993][17679847][critical][backtrace] [./source/server/backtrace.h:121] #40: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1033e282c]
[2025-07-09 10:48:08.994][17679847][critical][backtrace] [./source/server/backtrace.h:121] #41: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1033e2978]
[2025-07-09 10:48:08.995][17679847][critical][backtrace] [./source/server/backtrace.h:121] #42: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1033e0acc]
[2025-07-09 10:48:08.996][17679847][critical][backtrace] [./source/server/backtrace.h:121] #43: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x1033e3674]
[2025-07-09 10:48:08.998][17679847][critical][backtrace] [./source/server/backtrace.h:121] #44: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x1033e3564]
[2025-07-09 10:48:08.999][17679847][critical][backtrace] [./source/server/backtrace.h:121] #45: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x1033e3590]
[2025-07-09 10:48:09.000][17679847][critical][backtrace] [./source/server/backtrace.h:121] #46: std::__1::default_delete<>::operator()[abi:ne190102]() [0x1033e3cd0]
[2025-07-09 10:48:09.001][17679847][critical][backtrace] [./source/server/backtrace.h:121] #47: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1033e3c44]
[2025-07-09 10:48:09.002][17679847][critical][backtrace] [./source/server/backtrace.h:121] #48: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1033e3bd4]
[2025-07-09 10:48:09.004][17679847][critical][backtrace] [./source/server/backtrace.h:121] #49: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1033df314]
[2025-07-09 10:48:09.005][17679847][critical][backtrace] [./source/server/backtrace.h:121] #50: Envoy::Extensions::HttpFilters::OnDemand::OnDemandFilterConfig::~OnDemandFilterConfig() [0x1033e8de4]
[2025-07-09 10:48:09.006][17679847][critical][backtrace] [./source/server/backtrace.h:121] #51: Envoy::Extensions::HttpFilters::OnDemand::OnDemandFilterConfig::~OnDemandFilterConfig() [0x1033e07cc]
[2025-07-09 10:48:09.007][17679847][critical][backtrace] [./source/server/backtrace.h:121] #52: std::__1::__destroy_at[abi:ne190102]<>() [0x1033dd2f4]
[2025-07-09 10:48:09.008][17679847][critical][backtrace] [./source/server/backtrace.h:121] #53: std::__1::allocator_traits<>::destroy[abi:ne190102]<>() [0x1033dd2c8]
[2025-07-09 10:48:09.010][17679847][critical][backtrace] [./source/server/backtrace.h:121] #54: std::__1::__shared_ptr_emplace<>::__on_zero_shared_impl[abi:ne190102]<>() [0x1033dd298]
[2025-07-09 10:48:09.011][17679847][critical][backtrace] [./source/server/backtrace.h:121] #55: std::__1::__shared_ptr_emplace<>::__on_zero_shared() [0x1033dd08c]
[2025-07-09 10:48:09.012][17679847][critical][backtrace] [./source/server/backtrace.h:121] #56: std::__1::__shared_count::__release_shared[abi:ne190102]() [0x1022b2940]
[2025-07-09 10:48:09.013][17679847][critical][backtrace] [./source/server/backtrace.h:121] #57: std::__1::__shared_weak_count::__release_shared[abi:ne190102]() [0x1022b28e4]
[2025-07-09 10:48:09.014][17679847][critical][backtrace] [./source/server/backtrace.h:121] #58: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x10296c374]
[2025-07-09 10:48:09.016][17679847][critical][backtrace] [./source/server/backtrace.h:121] #59: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x10296c30c]
[2025-07-09 10:48:09.017][17679847][critical][backtrace] [./source/server/backtrace.h:121] #60: Envoy::Router::PerFilterConfigs::FilterConfig::~FilterConfig() [0x107564168]
./run_envoy.sh: line 17: 13114 Abort trap: 6           ~/tools/bin/envoy5 -c "$WORKSPACE_ROOT/envoy/bootstrap.yaml" --log-level debug --component-log-level dns:info
```
