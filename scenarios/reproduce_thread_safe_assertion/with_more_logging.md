Using envoy version built from commit 2c63725f4b2175bd74d04c64fa995d84311217f2 on Datadog/envoy-private
```
[2025-07-11 16:15:48.794][19558359][debug][upstream] [source/common/upstream/od_cds_api_impl.cc:39] odcds: creating subscription from config source
[2025-07-11 16:15:48.795][19558359][debug][router] [source/common/router/vhds.cc:108] vhds: loading new configuration: config_name=dynamic_route hash=15161633098654182640
[2025-07-11 16:15:48.795][19558359][debug][config] [source/extensions/config_subscription/grpc/delta_subscription_state.cc:289] Delta config for type.googleapis.com/envoy.config.route.v3.VirtualHost accepted with 1 resources added, 0 removed
[2025-07-11 16:15:48.796][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=main_thread, run_tid=19558359, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.796][19558462][error][misc] [source/common/grpc/async_client_impl.cc:333] gRPC AsyncStreamImpl::sendMessageRaw thread safety violation: dispatcher_name=main_thread, service=envoy.service.discovery.v3.AggregatedDiscoveryService, method=DeltaAggregatedResources, http_reset=false, waiting_to_delete=false
[2025-07-11 16:15:48.796][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=main_thread, run_tid=19558359, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.796][19558462][error][misc] [source/common/http/async_client_impl.cc:261] HTTP AsyncStreamImpl::sendData thread safety violation: dispatcher_name=main_thread, stream_info=xds_cluster, local_closed=false, remote_closed=false
[2025-07-11 16:15:48.796][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=main_thread, run_tid=19558359, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.796][19558462][critical][assert] [source/common/http/async_client_impl.cc:263] assert failure: dispatcher().isThreadSafe().
[2025-07-11 16:15:48.796][19558462][error][envoy_bug] [./source/common/common/assert.h:38] stacktrace for envoy bug
[2025-07-11 16:15:48.817][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #0 Envoy::Http::AsyncStreamImpl::sendData() [0x106feb4a8]
[2025-07-11 16:15:48.821][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #1 Envoy::Grpc::AsyncStreamImpl::sendMessageRaw() [0x106fd36a4]
[2025-07-11 16:15:48.822][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #2 Envoy::Grpc::Internal::sendMessageUntyped() [0x106fe02f4]
[2025-07-11 16:15:48.824][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #3 Envoy::Grpc::AsyncStream<>::sendMessage() [0x105c10d5c]
[2025-07-11 16:15:48.825][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #4 Envoy::Config::GrpcStream<>::sendMessage() [0x105c0b824]
[2025-07-11 16:15:48.827][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #5 Envoy::Config::NewGrpcMuxImpl::trySendDiscoveryRequests() [0x105bfc3c8]
[2025-07-11 16:15:48.828][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #6 Envoy::Config::NewGrpcMuxImpl::updateWatch() [0x105bff04c]
[2025-07-11 16:15:48.830][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #7 Envoy::Config::NewGrpcMuxImpl::removeWatch() [0x105bffd10]
[2025-07-11 16:15:48.831][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #8 Envoy::Config::NewGrpcMuxImpl::WatchImpl::remove() [0x105c22a5c]
[2025-07-11 16:15:48.833][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #9 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x105c229f8]
[2025-07-11 16:15:48.834][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #10 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x105c22950]
[2025-07-11 16:15:48.836][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #11 Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x105c2297c]
[2025-07-11 16:15:48.837][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #12 std::__1::default_delete<>::operator()[abi:ne190102]() [0x105bb5744]
[2025-07-11 16:15:48.838][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #13 std::__1::unique_ptr<>::reset[abi:ne190102]() [0x105bb56b8]
[2025-07-11 16:15:48.840][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #14 std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105bb5648]
[2025-07-11 16:15:48.841][19558462][error][envoy_bug] [./source/common/common/assert.h:43] #15 std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105bb31dc]
[2025-07-11 16:15:48.841][19558462][critical][backtrace] [./source/server/backtrace.h:129] Caught Abort trap: 6, suspect faulting address 0x19f9bd388
[2025-07-11 16:15:48.841][19558462][critical][backtrace] [./source/server/backtrace.h:113] Backtrace (use tools/stack_decode.py to get line numbers):
[2025-07-11 16:15:48.841][19558462][critical][backtrace] [./source/server/backtrace.h:114] Envoy version: 71bb03e3abc129d636e7715d633239017bf41773/1.34.1/Modified/DEBUG/BoringSSL
[2025-07-11 16:15:48.843][19558462][critical][backtrace] [./source/server/backtrace.h:121] #0: Envoy::SignalAction::sigHandler() [0x10843b31c]
[2025-07-11 16:15:48.843][19558462][critical][backtrace] [./source/server/backtrace.h:121] #1: _sigtramp [0x19fa30624]
[2025-07-11 16:15:48.843][19558462][critical][backtrace] [./source/server/backtrace.h:121] #2: pthread_kill [0x19f9f688c]
[2025-07-11 16:15:48.843][19558462][critical][backtrace] [./source/server/backtrace.h:121] #3: abort [0x19f8ffc60]
[2025-07-11 16:15:48.844][19558462][critical][backtrace] [./source/server/backtrace.h:121] #4: Envoy::Http::AsyncStreamImpl::sendData() [0x106feb4bc]
[2025-07-11 16:15:48.846][19558462][critical][backtrace] [./source/server/backtrace.h:121] #5: Envoy::Grpc::AsyncStreamImpl::sendMessageRaw() [0x106fd36a4]
[2025-07-11 16:15:48.847][19558462][critical][backtrace] [./source/server/backtrace.h:121] #6: Envoy::Grpc::Internal::sendMessageUntyped() [0x106fe02f4]
[2025-07-11 16:15:48.848][19558462][critical][backtrace] [./source/server/backtrace.h:121] #7: Envoy::Grpc::AsyncStream<>::sendMessage() [0x105c10d5c]
[2025-07-11 16:15:48.850][19558462][critical][backtrace] [./source/server/backtrace.h:121] #8: Envoy::Config::GrpcStream<>::sendMessage() [0x105c0b824]
[2025-07-11 16:15:48.851][19558462][critical][backtrace] [./source/server/backtrace.h:121] #9: Envoy::Config::NewGrpcMuxImpl::trySendDiscoveryRequests() [0x105bfc3c8]
[2025-07-11 16:15:48.853][19558462][critical][backtrace] [./source/server/backtrace.h:121] #10: Envoy::Config::NewGrpcMuxImpl::updateWatch() [0x105bff04c]
[2025-07-11 16:15:48.854][19558462][critical][backtrace] [./source/server/backtrace.h:121] #11: Envoy::Config::NewGrpcMuxImpl::removeWatch() [0x105bffd10]
[2025-07-11 16:15:48.855][19558462][critical][backtrace] [./source/server/backtrace.h:121] #12: Envoy::Config::NewGrpcMuxImpl::WatchImpl::remove() [0x105c22a5c]
[2025-07-11 16:15:48.856][19558462][critical][backtrace] [./source/server/backtrace.h:121] #13: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x105c229f8]
[2025-07-11 16:15:48.858][19558462][critical][backtrace] [./source/server/backtrace.h:121] #14: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x105c22950]
[2025-07-11 16:15:48.859][19558462][critical][backtrace] [./source/server/backtrace.h:121] #15: Envoy::Config::NewGrpcMuxImpl::WatchImpl::~WatchImpl() [0x105c2297c]
[2025-07-11 16:15:48.860][19558462][critical][backtrace] [./source/server/backtrace.h:121] #16: std::__1::default_delete<>::operator()[abi:ne190102]() [0x105bb5744]
[2025-07-11 16:15:48.862][19558462][critical][backtrace] [./source/server/backtrace.h:121] #17: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x105bb56b8]
[2025-07-11 16:15:48.863][19558462][critical][backtrace] [./source/server/backtrace.h:121] #18: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105bb5648]
[2025-07-11 16:15:48.865][19558462][critical][backtrace] [./source/server/backtrace.h:121] #19: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x105bb31dc]
[2025-07-11 16:15:48.866][19558462][critical][backtrace] [./source/server/backtrace.h:121] #20: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x105bb4dcc]
[2025-07-11 16:15:48.867][19558462][critical][backtrace] [./source/server/backtrace.h:121] #21: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x105bb5078]
[2025-07-11 16:15:48.869][19558462][critical][backtrace] [./source/server/backtrace.h:121] #22: Envoy::Config::GrpcSubscriptionImpl::~GrpcSubscriptionImpl() [0x105bb50a4]
[2025-07-11 16:15:48.870][19558462][critical][backtrace] [./source/server/backtrace.h:121] #23: std::__1::default_delete<>::operator()[abi:ne190102]() [0x102978a64]
[2025-07-11 16:15:48.871][19558462][critical][backtrace] [./source/server/backtrace.h:121] #24: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x1029789d8]
[2025-07-11 16:15:48.872][19558462][critical][backtrace] [./source/server/backtrace.h:121] #25: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x102978968]
[2025-07-11 16:15:48.874][19558462][critical][backtrace] [./source/server/backtrace.h:121] #26: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x10295c094]
[2025-07-11 16:15:48.875][19558462][critical][backtrace] [./source/server/backtrace.h:121] #27: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x10413e330]
[2025-07-11 16:15:48.876][19558462][critical][backtrace] [./source/server/backtrace.h:121] #28: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x10413dc44]
[2025-07-11 16:15:48.877][19558462][critical][backtrace] [./source/server/backtrace.h:121] #29: Envoy::Upstream::OdCdsApiImpl::~OdCdsApiImpl() [0x10413dc70]
[2025-07-11 16:15:48.878][19558462][critical][backtrace] [./source/server/backtrace.h:121] #30: std::__1::default_delete<>::operator()[abi:ne190102]() [0x10413ead4]
[2025-07-11 16:15:48.880][19558462][critical][backtrace] [./source/server/backtrace.h:121] #31: std::__1::__shared_ptr_pointer<>::__on_zero_shared() [0x10413e86c]
[2025-07-11 16:15:48.881][19558462][critical][backtrace] [./source/server/backtrace.h:121] #32: std::__1::__shared_count::__release_shared[abi:ne190102]() [0x1027e6940]
[2025-07-11 16:15:48.882][19558462][critical][backtrace] [./source/server/backtrace.h:121] #33: std::__1::__shared_weak_count::__release_shared[abi:ne190102]() [0x1027e68e4]
[2025-07-11 16:15:48.883][19558462][critical][backtrace] [./source/server/backtrace.h:121] #34: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x10413edec]
[2025-07-11 16:15:48.884][19558462][critical][backtrace] [./source/server/backtrace.h:121] #35: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x10413b300]
[2025-07-11 16:15:48.886][19558462][critical][backtrace] [./source/server/backtrace.h:121] #36: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x10612adb0]
[2025-07-11 16:15:48.887][19558462][critical][backtrace] [./source/server/backtrace.h:121] #37: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x10612abe0]
[2025-07-11 16:15:48.888][19558462][critical][backtrace] [./source/server/backtrace.h:121] #38: Envoy::Upstream::ClusterManagerImpl::OdCdsApiHandleImpl::~OdCdsApiHandleImpl() [0x10612ac0c]
[2025-07-11 16:15:48.889][19558462][critical][backtrace] [./source/server/backtrace.h:121] #39: std::__1::default_delete<>::operator()[abi:ne190102]() [0x103916920]
[2025-07-11 16:15:48.891][19558462][critical][backtrace] [./source/server/backtrace.h:121] #40: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x10391682c]
[2025-07-11 16:15:48.892][19558462][critical][backtrace] [./source/server/backtrace.h:121] #41: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x103916978]
[2025-07-11 16:15:48.893][19558462][critical][backtrace] [./source/server/backtrace.h:121] #42: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x103914acc]
[2025-07-11 16:15:48.894][19558462][critical][backtrace] [./source/server/backtrace.h:121] #43: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x103917674]
[2025-07-11 16:15:48.896][19558462][critical][backtrace] [./source/server/backtrace.h:121] #44: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x103917564]
[2025-07-11 16:15:48.897][19558462][critical][backtrace] [./source/server/backtrace.h:121] #45: Envoy::Extensions::HttpFilters::OnDemand::(anonymous namespace)::RdsCdsDecodeHeadersBehavior::~RdsCdsDecodeHeadersBehavior() [0x103917590]
[2025-07-11 16:15:48.898][19558462][critical][backtrace] [./source/server/backtrace.h:121] #46: std::__1::default_delete<>::operator()[abi:ne190102]() [0x103917cd0]
[2025-07-11 16:15:48.899][19558462][critical][backtrace] [./source/server/backtrace.h:121] #47: std::__1::unique_ptr<>::reset[abi:ne190102]() [0x103917c44]
[2025-07-11 16:15:48.901][19558462][critical][backtrace] [./source/server/backtrace.h:121] #48: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x103917bd4]
[2025-07-11 16:15:48.902][19558462][critical][backtrace] [./source/server/backtrace.h:121] #49: std::__1::unique_ptr<>::~unique_ptr[abi:ne190102]() [0x103913314]
[2025-07-11 16:15:48.903][19558462][critical][backtrace] [./source/server/backtrace.h:121] #50: Envoy::Extensions::HttpFilters::OnDemand::OnDemandFilterConfig::~OnDemandFilterConfig() [0x10391cde4]
[2025-07-11 16:15:48.904][19558462][critical][backtrace] [./source/server/backtrace.h:121] #51: Envoy::Extensions::HttpFilters::OnDemand::OnDemandFilterConfig::~OnDemandFilterConfig() [0x1039147cc]
[2025-07-11 16:15:48.906][19558462][critical][backtrace] [./source/server/backtrace.h:121] #52: std::__1::__destroy_at[abi:ne190102]<>() [0x1039112f4]
[2025-07-11 16:15:48.907][19558462][critical][backtrace] [./source/server/backtrace.h:121] #53: std::__1::allocator_traits<>::destroy[abi:ne190102]<>() [0x1039112c8]
[2025-07-11 16:15:48.908][19558462][critical][backtrace] [./source/server/backtrace.h:121] #54: std::__1::__shared_ptr_emplace<>::__on_zero_shared_impl[abi:ne190102]<>() [0x103911298]
[2025-07-11 16:15:48.909][19558462][critical][backtrace] [./source/server/backtrace.h:121] #55: std::__1::__shared_ptr_emplace<>::__on_zero_shared() [0x10391108c]
[2025-07-11 16:15:48.911][19558462][critical][backtrace] [./source/server/backtrace.h:121] #56: std::__1::__shared_count::__release_shared[abi:ne190102]() [0x1027e6940]
[2025-07-11 16:15:48.912][19558462][critical][backtrace] [./source/server/backtrace.h:121] #57: std::__1::__shared_weak_count::__release_shared[abi:ne190102]() [0x1027e68e4]
[2025-07-11 16:15:48.913][19558462][critical][backtrace] [./source/server/backtrace.h:121] #58: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x102ea0374]
[2025-07-11 16:15:48.914][19558462][critical][backtrace] [./source/server/backtrace.h:121] #59: std::__1::shared_ptr<>::~shared_ptr[abi:ne190102]() [0x102ea030c]
[2025-07-11 16:15:48.915][19558462][critical][backtrace] [./source/server/backtrace.h:121] #60: Envoy::Router::PerFilterConfigs::FilterConfig::~FilterConfig() [0x107a99cbc]
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=main_thread, run_tid=19558359, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_0, run_tid=19558453, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_1, run_tid=19558454, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_2, run_tid=19558455, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_3, run_tid=19558456, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_4, run_tid=19558457, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_5, run_tid=19558458, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_6, run_tid=19558459, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_7, run_tid=19558460, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=worker_8, run_tid=19558461, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=main_thread_guarddog_thread, run_tid=19558369, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
[2025-07-11 16:15:48.915][19558462][error][misc] [./source/common/event/dispatcher_impl.h:153] Dispatcher thread safety violation: name=workers_guarddog_thread, run_tid=19558370, current_tid=19558462, run_tid_empty=false, current_thread_role=worker
./run_envoy.sh: line 19: 80118 Abort trap: 6           $ENVOY -c "$WORKSPACE_ROOT/envoy/bootstrap.yaml" --log-level debug --component-log-level dns:info
```
