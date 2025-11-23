#!/usr/bin/env python3
"""
Phân tích logs từ Docker Compose để tìm vấn đề và trạng thái hệ thống
"""

import re
from collections import defaultdict
from datetime import datetime
from typing import Dict, List, Tuple

def parse_log_line(line: str) -> Dict:
    """Parse một dòng log"""
    result = {
        'service': None,
        'timestamp': None,
        'level': None,
        'message': None,
        'error': None,
        'raw': line.strip()
    }
    
    # Pattern cho log format: service_name | {"level":"...","msg":"..."}
    match = re.match(r'^(\S+)\s+\|\s+(.+)$', line)
    if not match:
        return result
    
    service = match.group(1)
    log_content = match.group(2)
    
    result['service'] = service
    
    # Parse JSON log (structured logging)
    if log_content.startswith('{'):
        try:
            import json
            log_data = json.loads(log_content)
            result['level'] = log_data.get('level', 'info')
            result['message'] = log_data.get('msg', '')
            result['timestamp'] = log_data.get('ts', 0)
            result['error'] = log_data.get('error', '')
            
            # Extract caller
            if 'caller' in log_data:
                result['caller'] = log_data['caller']
        except Exception as e:
            # If JSON parsing fails, treat as plain text
            result['message'] = log_content
    else:
        # Plain text log (nginx, couchdb, etc.)
        result['message'] = log_content
        # Try to detect level from plain text
        if 'error' in log_content.lower():
            result['level'] = 'error'
        elif 'warn' in log_content.lower():
            result['level'] = 'warn'
    
    return result

def analyze_logs(logs: List[str]) -> Dict:
    """Phân tích logs và trả về thống kê"""
    analysis = {
        'services': defaultdict(int),
        'errors': [],
        'warnings': [],
        'info': [],
        'fatal': [],
        'connection_errors': [],
        'startup_sequence': [],
        'health_checks': [],
        'service_status': {}
    }
    
    for line in logs:
        if not line.strip():
            continue
            
        parsed = parse_log_line(line)
        
        if parsed['service']:
            analysis['services'][parsed['service']] += 1
        
        level = (parsed.get('level') or 'info').lower()
        message = parsed.get('message') or ''
        error = parsed.get('error') or ''
        
        # Phân loại theo level
        if level == 'error':
            analysis['errors'].append({
                'service': parsed['service'],
                'message': message,
                'error': error,
                'caller': parsed.get('caller', ''),
                'raw': parsed.get('raw', '')
            })
        elif level == 'warn':
            analysis['warnings'].append({
                'service': parsed['service'],
                'message': message
            })
        elif level == 'fatal':
            analysis['fatal'].append({
                'service': parsed['service'],
                'message': message,
                'error': error,
                'raw': parsed.get('raw', '')
            })
        elif level == 'info':
            analysis['info'].append({
                'service': parsed['service'],
                'message': message
            })
        
        # Tìm connection errors
        full_text = (message + ' ' + error).lower()
        if 'connection refused' in full_text or 'connect() failed' in full_text or 'dial error' in full_text:
            analysis['connection_errors'].append({
                'service': parsed['service'],
                'message': message,
                'error': error,
                'raw': parsed.get('raw', '')
            })
        
        # Tìm startup sequence
        if any(keyword in message.lower() for keyword in ['starting', 'started', 'connected', 'initialized']):
            analysis['startup_sequence'].append({
                'service': parsed['service'],
                'message': message,
                'timestamp': parsed.get('timestamp', 0)
            })
        
        # Health checks
        if 'health' in message.lower() or '/health' in message:
            analysis['health_checks'].append({
                'service': parsed['service'],
                'message': message,
                'status': '200' if 'status":200' in line else 'unknown'
            })
    
    return analysis

def print_analysis(analysis: Dict):
    """In kết quả phân tích"""
    print("=" * 80)
    print("PHÂN TÍCH LOGS HỆ THỐNG IBN")
    print("=" * 80)
    
    print("\n📊 THỐNG KÊ SERVICES:")
    print("-" * 80)
    for service, count in sorted(analysis['services'].items(), key=lambda x: x[1], reverse=True):
        print(f"  {service:30s} : {count:4d} logs")
    
    print("\n❌ LỖI (ERRORS):")
    print("-" * 80)
    if analysis['errors']:
        for err in analysis['errors']:
            print(f"  [{err['service']}] {err['message']}")
            if err['error']:
                print(f"    Error: {err['error']}")
            if err.get('caller'):
                print(f"    Caller: {err['caller']}")
    else:
        print("  ✅ Không có lỗi")
    
    print("\n⚠️  CẢNH BÁO (WARNINGS):")
    print("-" * 80)
    if analysis['warnings']:
        for warn in analysis['warnings'][:10]:  # Limit to 10
            print(f"  [{warn['service']}] {warn['message']}")
    else:
        print("  ✅ Không có cảnh báo")
    
    print("\n💀 LỖI NGHIÊM TRỌNG (FATAL):")
    print("-" * 80)
    if analysis['fatal']:
        for fatal in analysis['fatal']:
            print(f"  [{fatal['service']}]")
            print(f"    Message: {fatal['message']}")
            if fatal['error']:
                print(f"    Error: {fatal['error']}")
    else:
        print("  ✅ Không có lỗi nghiêm trọng")
    
    print("\n🔌 LỖI KẾT NỐI (CONNECTION ERRORS):")
    print("-" * 80)
    if analysis['connection_errors']:
        for i, conn_err in enumerate(analysis['connection_errors'], 1):
            print(f"\n  [{i}] Service: {conn_err['service']}")
            if conn_err['message']:
                print(f"    Message: {conn_err['message'][:100]}...")
            if conn_err['error']:
                print(f"    Error: {conn_err['error'][:150]}...")
            # Extract connection details
            raw = conn_err.get('raw', '')
            if '172.21.0.' in raw or '7051' in raw or '5432' in raw:
                if '7051' in raw:
                    print(f"    → Không thể kết nối đến Fabric Peer (port 7051)")
                elif '5432' in raw:
                    print(f"    → Không thể kết nối đến PostgreSQL (port 5432)")
                elif '8080' in raw:
                    print(f"    → Nginx không thể kết nối đến API Gateway (port 8080)")
    else:
        print("  ✅ Không có lỗi kết nối")
    
    print("\n🚀 QUÁ TRÌNH KHỞI ĐỘNG (STARTUP SEQUENCE):")
    print("-" * 80)
    startup_by_service = defaultdict(list)
    for startup in analysis['startup_sequence']:
        startup_by_service[startup['service']].append(startup['message'])
    
    for service, messages in startup_by_service.items():
        print(f"  [{service}]")
        for msg in messages[-5:]:  # Last 5 messages
            print(f"    - {msg}")
    
    print("\n💚 HEALTH CHECKS:")
    print("-" * 80)
    if analysis['health_checks']:
        health_by_service = defaultdict(list)
        for health in analysis['health_checks']:
            health_by_service[health['service']].append(health['status'])
        
        for service, statuses in health_by_service.items():
            success = sum(1 for s in statuses if s == '200')
            total = len(statuses)
            print(f"  [{service}] {success}/{total} successful checks")
    else:
        print("  ⚠️  Không có health check logs")
    
    print("\n" + "=" * 80)
    print("TÓM TẮT:")
    print("-" * 80)
    print(f"  Tổng số services: {len(analysis['services'])}")
    print(f"  Tổng số lỗi: {len(analysis['errors'])}")
    print(f"  Tổng số cảnh báo: {len(analysis['warnings'])}")
    print(f"  Tổng số lỗi nghiêm trọng: {len(analysis['fatal'])}")
    print(f"  Tổng số lỗi kết nối: {len(analysis['connection_errors'])}")
    print("=" * 80)

def main():
    # Đọc logs từ terminal selection hoặc file
    logs = """
api-gateway-1          | {"level":"info","ts":1763352217.7894428,"caller":"metrics/service.go:102","msg":"Getting transaction metrics","channel":"ibnchannel"}
api-gateway-1          | {"level":"info","ts":1763352217.7911122,"caller":"metrics/service.go:192","msg":"Getting block metrics","channel":"ibnchannel"}
api-gateway-1          | {"level":"info","ts":1763352217.7911325,"caller":"explorer/service.go:172","msg":"Listing blocks","channel":"ibnchannel","limit":10000,"offset":0}
api-gateway-1          | {"level":"info","ts":1763352217.7919984,"caller":"metrics/service.go:255","msg":"Getting performance metrics"}
api-gateway-1          | {"level":"info","ts":1763352217.8408363,"caller":"metrics/service.go:332","msg":"Getting peer metrics"}
api-gateway-1          | {"level":"info","ts":1763352217.84101,"caller":"explorer/service.go:172","msg":"Listing blocks","channel":"ibnchannel","limit":10,"offset":0}
api-gateway-1          | {"level":"info","ts":1763352219.2955513,"caller":"server/main.go:318","msg":"Shutting down server..."}
api-gateway-1          | {"level":"info","ts":1763352219.2956862,"caller":"fabric/gateway.go:178","msg":"Fabric Gateway connection closed"}
api-gateway-1          | {"level":"info","ts":1763352219.295815,"caller":"cache/redis.go:167","msg":"Redis connection closed"}
api-gateway-1          | {"level":"info","ts":1763352219.2962372,"caller":"server/main.go:343","msg":"All services closed"}
api-gateway-1          | {"level":"info","ts":1763352219.296357,"caller":"server/main.go:345","msg":"Server exited gracefully"}
api-gateway-1          | {"level":"info","ts":1763352219.2963822,"caller":"indexer/service.go:70","msg":"Stopping block indexer service"}
api-gateway-1          | {"level":"info","ts":1763352219.2964,"caller":"fabric/gateway.go:178","msg":"Fabric Gateway connection closed"}
api-gateway-1          | {"level":"info","ts":1763352222.1308389,"caller":"server/main.go:87","msg":"Starting IBN API Gateway","version":"1.0.0","environment":"production"}
api-gateway-1          | {"level":"fatal","ts":1763352222.1497364,"caller":"server/main.go:95","msg":"Failed to connect to database","error":"failed to ping database: failed to connect to `host=postgres user=gateway database=ibn_gateway`: dial error (dial tcp 172.21.0.14:5432: connect: connection refused)","stacktrace":"main.main\n\t/app/cmd/server/main.go:95\nruntime.main\n\t/usr/local/go/src/runtime/proc.go:272"}
api-gateway-1          | {"level":"info","ts":1763352224.3705986,"caller":"server/main.go:87","msg":"Starting IBN API Gateway","version":"1.0.0","environment":"production"}
api-gateway-1          | {"level":"info","ts":1763352224.3877275,"caller":"server/main.go:98","msg":"Connected to PostgreSQL"}
api-gateway-1          | {"level":"info","ts":1763352224.389145,"caller":"cache/redis.go:36","msg":"Connected to Redis","address":"redis:6379"}
api-gateway-1          | {"level":"info","ts":1763352224.3899748,"caller":"fabric/gateway.go:113","msg":"Connected to Fabric Gateway","channel":"ibnchannel","chaincode":"teaTraceCC","mspId":"Org1MSP"}
api-gateway-1          | {"level":"info","ts":1763352224.390047,"caller":"indexer/service.go:53","msg":"Starting block indexer service"}
api-gateway-1          | {"level":"info","ts":1763352224.3900816,"caller":"server/main.go:228","msg":"Block indexer started successfully"}
api-gateway-1          | {"level":"info","ts":1763352224.3901243,"caller":"indexer/service.go:78","msg":"Indexing historical blocks from transactions","channel":"ibnchannel"}
api-gateway-1          | {"level":"info","ts":1763352224.390702,"caller":"server/main.go:308","msg":"API Gateway started successfully","address":"0.0.0.0:8080","swagger":"http://0.0.0.0:8080/swagger/index.html"}
api-gateway-1          | {"level":"info","ts":1763352224.39079,"caller":"server/main.go:302","msg":"Starting HTTP server","address":"0.0.0.0:8080"}
api-gateway-1          | {"level":"error","ts":1763352224.3913186,"caller":"indexer/service.go:223","msg":"Failed to create block event listener","channel":"ibnchannel","error":"rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial tcp 172.21.0.26:7051: connect: connection refused\"","stacktrace":"github.com/ibn-network/api-gateway/internal/services/indexer.(*Service).listenToChannelBlockEvents\n\t/app/internal/services/indexer/service.go:223"}
api-gateway-1          | {"level":"info","ts":1763352229.4033434,"caller":"middleware/logger.go:56","msg":"HTTP request","correlation_id":"dcb14197-e9ce-43e1-9390-9a29942472d4","method":"GET","path":"/health","remote_addr":"[::1]:44226","status":200,"duration":0.00106619,"user_agent":"Wget"}
api-gateway-2          | {"level":"info","ts":1763352246.0748937,"caller":"metrics/service.go:349","msg":"Getting metrics summary","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352246.074976,"caller":"metrics/service.go:102","msg":"Getting transaction metrics","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352246.0778804,"caller":"metrics/service.go:192","msg":"Getting block metrics","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352246.0779269,"caller":"explorer/service.go:172","msg":"Listing blocks","channel":"ibnchannel","limit":10000,"offset":0}
api-gateway-2          | {"level":"info","ts":1763352246.0796819,"caller":"metrics/service.go:255","msg":"Getting performance metrics"}
api-gateway-2          | {"level":"info","ts":1763352246.172074,"caller":"metrics/service.go:332","msg":"Getting peer metrics"}
api-gateway-2          | {"level":"info","ts":1763352246.1724463,"caller":"explorer/service.go:172","msg":"Listing blocks","channel":"ibnchannel","limit":10,"offset":0}
couchdb0               | [notice] 2025-11-17T04:04:09.169298Z nonode@nohost <0.1086.0> 655de37663 localhost:5984 127.0.0.1 admin GET /_up 200 ok 2
couchdb2               | [notice] 2025-11-17T04:04:09.274739Z nonode@nohost <0.1086.0> 3ec07593b7 localhost:5984 127.0.0.1 admin GET /_up 200 ok 2
couchdb1               | [notice] 2025-11-17T04:04:09.398566Z nonode@nohost <0.1086.0> ed674454c9 localhost:5984 127.0.0.1 admin GET /_up 200 ok 2
node-exporter-proxy    | 172.21.0.8 - - [17/Nov/2025:04:04:13 +0000] "GET /metrics HTTP/1.1" 200 8953 "-" "Prometheus/2.45.0"
api-gateway-3          | {"level":"info","ts":1763352253.9467509,"caller":"indexer/service.go:78","msg":"Indexing historical blocks from transactions","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352254.0501363,"caller":"indexer/service.go:78","msg":"Indexing historical blocks from transactions","channel":"ibnchannel"}
api-gateway-1          | {"level":"info","ts":1763352254.3997643,"caller":"indexer/service.go:78","msg":"Indexing historical blocks from transactions","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352256.0670896,"caller":"metrics/service.go:349","msg":"Getting metrics summary","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352256.067364,"caller":"metrics/service.go:102","msg":"Getting transaction metrics","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352256.0740535,"caller":"metrics/service.go:192","msg":"Getting block metrics","channel":"ibnchannel"}
api-gateway-2          | {"level":"info","ts":1763352256.0741014,"caller":"explorer/service.go:172","msg":"Listing blocks","channel":"ibnchannel","limit":10000,"offset":0}
api-gateway-2          | {"level":"info","ts":1763352256.0805638,"caller":"metrics/service.go:255","msg":"Getting performance metrics"}
api-gateway-2          | {"level":"info","ts":1763352256.1841671,"caller":"metrics/service.go:332","msg":"Getting peer metrics"}
api-gateway-2          | {"level":"info","ts":1763352256.1845617,"caller":"explorer/service.go:172","msg":"Listing blocks","channel":"ibnchannel","limit":10,"offset":0}
api-gateway-nginx      | 2025/11/17 04:04:17 [error] 24#24: *11 connect() failed (111: Connection refused) while connecting to upstream, client: 127.0.0.1, server: localhost, request: "GET /health HTTP/1.1", upstream: "http://172.21.0.15:8080/health", host: "127.0.0.1"
api-gateway-2          | {"level":"info","ts":1763352257.946215,"caller":"middleware/logger.go:56","msg":"HTTP request","correlation_id":"a0d07b56-1b84-4b99-adaf-3f7e8df8409d","method":"GET","path":"/health","remote_addr":"172.21.0.7:52908","status":200,"duration":0.001270798,"user_agent":"Wget"}
api-gateway-3          | {"level":"info","ts":1763352258.9612112,"caller":"middleware/logger.go:56","msg":"HTTP request","correlation_id":"849d53f2-7696-43c6-9acb-7f5643508223","method":"GET","path":"/health","remote_addr":"[::1]:58720","status":200,"duration":0.001117759,"user_agent":"Wget"}
api-gateway-2          | {"level":"info","ts":1763352258.9990501,"caller":"middleware/logger.go:56","msg":"HTTP request","correlation_id":"ca30490b-e64f-4189-93fe-d043a84ffcbf","method":"GET","path":"/health","remote_addr":"[::1]:58732","status":200,"duration":0.001167165,"user_agent":"Wget"}
couchdb0               | [notice] 2025-11-17T04:04:09.250827Z nonode@nohost <0.1169.0> d309e8dd2a localhost:5984 127.0.0.1 admin GET /_up 200 ok 2
couchdb2               | [notice] 2025-11-17T04:04:09.345740Z nonode@nohost <0.1169.0> 7f9bcd003e localhost:5984 127.0.0.1 admin  GET /_up 200 ok 2
couchdb1               | [notice] 2025-11-17T04:04:09.482872Z nonode@nohost <0.1169.0> a8e6523e47 localhost:5984 127.0.0.1 admin GET /_up 200 ok 2
api-gateway-1          | {"level":"info","ts":1763352259.4862156,"caller":"middleware/logger.go:56","msg":"HTTP request","correlation_id":"1811a07b-f837-44a4-9b7c-3bea176a8c1f","method":"GET","path":"/health","remote_addr":"[::1]:58736","status":200,"duration":0.001117759,"user_agent":"Wget"}
api-gateway-1          | {"level":"info","ts":1763351398.0957465,"caller":"metrics/service.go:332","msg":"Getting peer metrics"}
""".strip().split('\n')
    
    analysis = analyze_logs(logs)
    print_analysis(analysis)
    
    # Đưa ra khuyến nghị
    print("\n💡 KHUYẾN NGHỊ:")
    print("-" * 80)
    
    if analysis['fatal']:
        print("  ⚠️  CÓ LỖI NGHIÊM TRỌNG:")
        for fatal in analysis['fatal']:
            if 'database' in fatal['error'].lower():
                print("    - Database chưa sẵn sàng khi API Gateway khởi động")
                print("      → Cần thêm healthcheck và depends_on trong docker-compose")
    
    if analysis['connection_errors']:
        print("  ⚠️  LỖI KẾT NỐI:")
        for conn_err in analysis['connection_errors']:
            if 'fabric' in conn_err['error'].lower() or '7051' in conn_err['error']:
                print("    - Không thể kết nối đến Fabric peer (port 7051)")
                print("      → Kiểm tra peer có đang chạy không")
                print("      → Kiểm tra network configuration")
            elif 'postgres' in conn_err['error'].lower() or '5432' in conn_err['error']:
                print("    - Không thể kết nối đến PostgreSQL")
                print("      → Kiểm tra postgres service có healthy không")
    
    if any('nginx' in err.get('service', '') for err in analysis['connection_errors']):
        print("    - Nginx không thể kết nối đến upstream")
        print("      → Kiểm tra load balancer configuration")
        print("      → Đảm bảo tất cả API Gateway instances đang chạy")
    
    if len(analysis['health_checks']) > 0:
        print("  ✅ Health checks đang hoạt động")
    
    print("\n")

if __name__ == '__main__':
    main()

