* Use debug version of envoy with assertions
* run the control-plane with `./scripts/build_and_run_control_plane.sh`
* run envoy with `./scripts/run_envoy.sh`
* then to update vhost config run `go run cli/*.go -config examples/dynamic_vhost_cluster1.json -action update`
* make a request to the vhost you just added `./make_request.sh 10002`
* change the vhost defintion `go run cli/*.go  -config examples/dynamic_vhost_cluster2.json -action update`

It should have the following assertion error

```
[2025-07-09 11:23:02.010][17724758][debug][config] [source/extensions/config_subscription/grpc/new_grpc_mux_impl.cc:158] Received DeltaDiscoveryResponse for type.googleapis.com/envoy.config.route.v3.VirtualHost at version 4
[2025-07-09 11:23:02.010][17724758][debug][upstream] [source/common/upstream/od_cds_api_impl.cc:39] odcds: creating subscription from config source
[2025-07-09 11:23:02.011][17724758][debug][router] [source/common/router/vhds.cc:108] vhds: loading new configuration: config_name=dynamic_route hash=15161633098654182640
[2025-07-09 11:23:02.011][17724758][debug][config] [source/extensions/config_subscription/grpc/delta_subscription_state.cc:289] Delta config for type.googleapis.com/envoy.config.route.v3.VirtualHost accepted with 1 resources added, 0 removed
[2025-07-09 11:23:02.011][17724832][critical][assert] [source/common/http/async_client_impl.cc:257] assert failure: dispatcher().isThreadSafe().
[2025-07-09 11:23:02.011][17724832][error][envoy_bug] [./source/common/common/assert.h:38] stacktrace for envoy bug
[2025-07-09 11:23:02.029][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #0 Envoy::Http::AsyncStreamImpl::sendData() [0x108c5ddc0]
[2025-07-09 11:23:02.032][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #1 Envoy::Grpc::AsyncStreamImpl::sendMessageRaw() [0x108c467b0]
[2025-07-09 11:23:02.034][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #2 Envoy::Grpc::Internal::sendMessageUntyped() [0x108c52ea0]
[2025-07-09 11:23:02.035][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #3 Envoy::Grpc::AsyncStream<>::sendMessage() [0x107884d5c]
[2025-07-09 11:23:02.036][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #4 Envoy::Config::GrpcStream<>::sendMessage() [0x10787f824]
[2025-07-09 11:23:02.038][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #5 Envoy::Config::NewGrpcMuxImpl::trySendDiscoveryRequests() [0x1078703c8]
[2025-07-09 11:23:02.039][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #6 Envoy::Config::NewGrpcMuxImpl::updateWatch() [0x10787304c]
[2025-07-09 11:23:02.040][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #7 Envoy::Config::NewGrpcMuxImpl::removeWatch() [0x107873d10]
[2025-07-09 11:23:02.042][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #8 Envoy::Config::NewGrpcMuxImpl::WatchImpl::remove() [0x107896a5c]
[2025-07-09 11:23:02.043][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #9 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1078969f8]
[2025-07-09 11:23:02.045][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #10 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x107896950]
[2025-07-09 11:23:02.046][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #11 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x10789697c]
[2025-07-09 11:23:02.047][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #12 std::__1::default_delete<>::operator()[abi:ne190102]() [0x107829744]
[2025-07-09 11:23:02.049][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #13 std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1078296b8]
[2025-07-09 11:23:02.050][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #14 std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x107829648]
[2025-07-09 11:23:02.052][17724832][error][envoy_bug] [./source/common/common/assert.h:43] #15 std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1078271dc]
[2025-07-09 11:23:02.052][17724832][critical][backtrace] [./source/server/backtrace.h:129] Caught Abort trap: 6, suspect faulting address 0x19f9bd388
[2025-07-09 11:23:02.052][17724832][critical][backtrace] [./source/server/backtrace.h:113] Backtrace (use tools/stack_decode.py to get line numbers):
[2025-07-09 11:23:02.052][17724832][critical][backtrace] [./source/server/backtrace.h:114] Envoy version: a17185b6617c6a631607ba32cda4a2d3c32f9b3b/1.34.1/Modified/DEBUG/BoringSSL
[2025-07-09 11:23:02.053][17724832][critical][backtrace] [./source/server/backtrace.h:121] #0: Envoy::SignalAction::sigHandler() [0x10a0ad7c8]
[2025-07-09 11:23:02.053][17724832][critical][backtrace] [./source/server/backtrace.h:121] #1: _sigtramp [0x19fa30624]
[2025-07-09 11:23:02.053][17724832][critical][backtrace] [./source/server/backtrace.h:121] #2: pthread_kill [0x19f9f688c]
[2025-07-09 11:23:02.053][17724832][critical][backtrace] [./source/server/backtrace.h:121] #3: abort [0x19f8ffc60]
[2025-07-09 11:23:02.055][17724832][critical][backtrace] [./source/server/backtrace.h:121] #4: Envoy::Http::AsyncStreamImpl::sendData() [0x108c5ddd4]
[2025-07-09 11:23:02.056][17724832][critical][backtrace] [./source/server/backtrace.h:121] #5: Envoy::Grpc::AsyncStreamImpl::sendMessageRaw() [0x108c467b0]
[2025-07-09 11:23:02.057][17724832][critical][backtrace] [./source/server/backtrace.h:121] #6: Envoy::Grpc::Internal::sendMessageUntyped() [0x108c52ea0]
[2025-07-09 11:23:02.059][17724832][critical][backtrace] [./source/server/backtrace.h:121] #7: Envoy::Grpc::AsyncStream<>::sendMessage() [0x107884d5c]
[2025-07-09 11:23:02.060][17724832][critical][backtrace] [./source/server/backtrace.h:121] #8: Envoy::Config::GrpcStream<>::sendMessage() [0x10787f824]
[2025-07-09 11:23:02.061][17724832][critical][backtrace] [./source/server/backtrace.h:121] #9: Envoy::Config::NewGrpcMuxImpl::trySendDiscoveryRequests() [0x1078703c8]
[2025-07-09 11:23:02.063][17724832][critical][backtrace] [./source/server/backtrace.h:121] #10: Envoy::Config::NewGrpcMuxImpl::updateWatch() [0x10787304c]
[2025-07-09 11:23:02.064][17724832][critical][backtrace] [./source/server/backtrace.h:121] #11: Envoy::Config::NewGrpcMuxImpl::removeWatch() [0x107873d10]
[2025-07-09 11:23:02.065][17724832][critical][backtrace] [./source/server/backtrace.h:121] #12: Envoy::Config::NewGrpcMuxImpl::WatchImpl::remove() [0x107896a5c]
[2025-07-09 11:23:02.067][17724832][critical][backtrace] [./source/server/backtrace.h:121] #13: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x1078969f8]
[2025-07-09 11:23:02.068][17724832][critical][backtrace] [./source/server/backtrace.h:121] #14: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x107896950]
[2025-07-09 11:23:02.069][17724832][critical][backtrace] [./source/server/backtrace.h:121] #15: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x10789697c]
[2025-07-09 11:23:02.071][17724832][critical][backtrace] [./source/server/backtrace.h:121] #16: std::__1::default_delete<>::operator()[abi:ne190102]() [0x107829744]
[2025-07-09 11:23:02.072][17724832][critical][backtrace] [./source/server/backtrace.h:121] #17: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1078296b8]
[2025-07-09 11:23:02.073][17724832][critical][backtrace] [./source/server/backtrace.h:121] #18: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x107829648]
[2025-07-09 11:23:02.075][17724832][critical][backtrace] [./source/server/backtrace.h:121] #19: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1078271dc]
[2025-07-09 11:23:02.076][17724832][critical][backtrace] [./source/server/backtrace.h:121] #20: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x107828dcc]
[2025-07-09 11:23:02.077][17724832][critical][backtrace] [./source/server/backtrace.h:121] #21: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x107829078]
[2025-07-09 11:23:02.079][17724832][critical][backtrace] [./source/server/backtrace.h:121] #22: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x1078290a4]
[2025-07-09 11:23:02.080][17724832][critical][backtrace] [./source/server/backtrace.h:121] #23: std::__1::default_delete<>::operator()[abi:ne190102]() [0x1045eca64]
[2025-07-09 11:23:02.081][17724832][critical][backtrace] [./source/server/backtrace.h:121] #24: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1045ec9d8]
[2025-07-09 11:23:02.082][17724832][critical][backtrace] [./source/server/backtrace.h:121] #25: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1045ec968]
[2025-07-09 11:23:02.083][17724832][critical][backtrace] [./source/server/backtrace.h:121] #26: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x1045d0094]
[2025-07-09 11:23:02.085][17724832][critical][backtrace] [./source/server/backtrace.h:121] #27: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x105db2330]
[2025-07-09 11:23:02.086][17724832][critical][backtrace] [./source/server/backtrace.h:121] #28: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x105db1c44]
[2025-07-09 11:23:02.087][17724832][critical][backtrace] [./source/server/backtrace.h:121] #29: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x105db1c70]
[2025-07-09 11:23:02.088][17724832][critical][backtrace] [./source/server/backtrace.h:121] #30: std::__1::default_delete<>::operator()[abi:ne190102]() [0x105db2ad4]
[2025-07-09 11:23:02.090][17724832][critical][backtrace] [./source/server/backtrace.h:121] #31: std::__1::__shared_ptr_pointer<>::__on_zero_shared() [0x105db286c]
[2025-07-09 11:23:02.091][17724832][critical][backtrace] [./source/server/backtrace.h:121] #32: std::__1::__shared_count::__release_shared[abi:ne190102]() [0x10445a940]
[2025-07-09 11:23:02.092][17724832][critical][backtrace] [./source/server/backtrace.h:121] #33: std::__1::__shared_weak_count::__release_shared[abi:ne190102]() [0x10445a8e4]
[2025-07-09 11:23:02.093][17724832][critical][backtrace] [./source/server/backtrace.h:121] #34: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x105db2dec]
[2025-07-09 11:23:02.094][17724832][critical][backtrace] [./source/server/backtrace.h:121] #35: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x105daf300]
[2025-07-09 11:23:02.096][17724832][critical][backtrace] [./source/server/backtrace.h:121] #36: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x107d9e528]
[2025-07-09 11:23:02.097][17724832][critical][backtrace] [./source/server/backtrace.h:121] #37: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x107d9e358]
[2025-07-09 11:23:02.098][17724832][critical][backtrace] [./source/server/backtrace.h:121] #38: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x107d9e384]
[2025-07-09 11:23:02.100][17724832][critical][backtrace] [./source/server/backtrace.h:121] #39: std::__1::default_delete<>::operator()[abi:ne190102]() [0x10558a920]
[2025-07-09 11:23:02.101][17724832][critical][backtrace] [./source/server/backtrace.h:121] #40: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x10558a82c]
[2025-07-09 11:23:02.102][17724832][critical][backtrace] [./source/server/backtrace.h:121] #41: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x10558a978]
[2025-07-09 11:23:02.103][17724832][critical][backtrace] [./source/server/backtrace.h:121] #42: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105588acc]
[2025-07-09 11:23:02.105][17724832][critical][backtrace] [./source/server/backtrace.h:121] #43: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x10558b674]
[2025-07-09 11:23:02.106][17724832][critical][backtrace] [./source/server/backtrace.h:121] #44: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x10558b564]
[2025-07-09 11:23:02.107][17724832][critical][backtrace] [./source/server/backtrace.h:121] #45: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x10558b590]
[2025-07-09 11:23:02.108][17724832][critical][backtrace] [./source/server/backtrace.h:121] #46: std::__1::default_delete<>::operator()[abi:ne190102]() [0x10558bcd0]
[2025-07-09 11:23:02.110][17724832][critical][backtrace] [./source/server/backtrace.h:121] #47: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x10558bc44]
[2025-07-09 11:23:02.111][17724832][critical][backtrace] [./source/server/backtrace.h:121] #48: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x10558bbd4]
[2025-07-09 11:23:02.112][17724832][critical][backtrace] [./source/server/backtrace.h:121] #49: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105587314]
[2025-07-09 11:23:02.113][17724832][critical][backtrace] [./source/server/backtrace.h:121] #50: Envoy::Extensions::HttpFilters::OnDemand::OnDemandFilterConfig::~OnDemandFilterConfig() [0x105590de4]
[2025-07-09 11:23:02.114][17724832][critical][backtrace] [./source/server/backtrace.h:121] #51: Envoy::Extensions::HttpFilters::OnDemand::OnDemandFilterConfig::~OnDemandFilterConfig() [0x1055887cc]
[2025-07-09 11:23:02.116][17724832][critical][backtrace] [./source/server/backtrace.h:121] #52: std::__1::__destroy_at[abi:ne190102]<>() [0x1055852f4]
[2025-07-09 11:23:02.117][17724832][critical][backtrace] [./source/server/backtrace.h:121] #53: std::__1::allocator_traits<>::destroy[abi:ne190102]<>() [0x1055852c8]
[2025-07-09 11:23:02.118][17724832][critical][backtrace] [./source/server/backtrace.h:121] #54: std::__1::__shared_ptr_emplace<>::__on_zero_shared_impl[abi:ne190102]<>() [0x105585298]
[2025-07-09 11:23:02.119][17724832][critical][backtrace] [./source/server/backtrace.h:121] #55: std::__1::__shared_ptr_emplace<>::__on_zero_shared() [0x10558508c]
[2025-07-09 11:23:02.120][17724832][critical][backtrace] [./source/server/backtrace.h:121] #56: std::__1::__shared_count::__release_shared[abi:ne190102]() [0x10445a940]
[2025-07-09 11:23:02.122][17724832][critical][backtrace] [./source/server/backtrace.h:121] #57: std::__1::__shared_weak_count::__release_shared[abi:ne190102]() [0x10445a8e4]
[2025-07-09 11:23:02.123][17724832][critical][backtrace] [./source/server/backtrace.h:121] #58: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x104b14374]
[2025-07-09 11:23:02.124][17724832][critical][backtrace] [./source/server/backtrace.h:121] #59: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x104b1430c]
[2025-07-09 11:23:02.125][17724832][critical][backtrace] [./source/server/backtrace.h:121] #60: Envoy::Router::PerFilterConfigs::FilterConfig::~FilterConfig() [0x10970c168]
./run_envoy.sh: line 17: 22223 Abort trap: 6           ~/tools/bin/envoy5 -c "$WORKSPACE_ROOT/envoy/bootstrap.yaml" --log-level debug --component-log-level dns:info
```
