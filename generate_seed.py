import random
import uuid
import datetime
import hashlib

def generate_tx_id():
    return hashlib.sha256(str(uuid.uuid4()).encode()).hexdigest()

def generate_block_hash():
    return hashlib.sha256(str(uuid.uuid4()).encode()).hexdigest()

def get_timestamp(days_ago):
    return (datetime.datetime.now() - datetime.timedelta(days=days_ago)).isoformat()

batches = []
packages = []

sql_statements = []

# Create Batches
for i in range(5):
    tx_id = generate_tx_id()
    batch_id = f"BATCH_{uuid.uuid4().hex[:8].upper()}"
    batches.append(batch_id)
    block_number = random.randint(10, 100)
    block_hash = generate_block_hash()
    timestamp = get_timestamp(random.randint(5, 10))
    
    args = f'["{batch_id}", "Tea Batch {i+1}", "Organic Green Tea", "1000kg"]'
    
    sql = f"""INSERT INTO transactions (id, tx_id, channel_name, chaincode_name, function_name, args, status, block_number, block_hash, timestamp) 
    VALUES (gen_random_uuid(), '{tx_id}', 'ibnchannel', 'teatrace', 'CreateBatch', '{args}', 'VALID', {block_number}, '{block_hash}', '{timestamp}');"""
    sql_statements.append(sql)

# Create Packages
for i in range(10):
    tx_id = generate_tx_id()
    package_id = f"PKG_{uuid.uuid4().hex[:8].upper()}"
    packages.append(package_id)
    batch_id = random.choice(batches)
    block_number = random.randint(101, 200)
    block_hash = generate_block_hash()
    timestamp = get_timestamp(random.randint(3, 5))
    
    args = f'["{package_id}", "{batch_id}", "Premium Tea Package", "500g"]'
    
    sql = f"""INSERT INTO transactions (id, tx_id, channel_name, chaincode_name, function_name, args, status, block_number, block_hash, timestamp) 
    VALUES (gen_random_uuid(), '{tx_id}', 'ibnchannel', 'teatrace', 'CreatePackage', '{args}', 'VALID', {block_number}, '{block_hash}', '{timestamp}');"""
    sql_statements.append(sql)

# Update Packages
for i in range(8):
    tx_id = generate_tx_id()
    package_id = random.choice(packages)
    block_number = random.randint(201, 300)
    block_hash = generate_block_hash()
    timestamp = get_timestamp(random.randint(1, 2))
    
    args = f'["{package_id}", "PROCESSING", "Drying tea leaves"]'
    
    sql = f"""INSERT INTO transactions (id, tx_id, channel_name, chaincode_name, function_name, args, status, block_number, block_hash, timestamp) 
    VALUES (gen_random_uuid(), '{tx_id}', 'ibnchannel', 'teatrace', 'UpdatePackage', '{args}', 'VALID', {block_number}, '{block_hash}', '{timestamp}');"""
    sql_statements.append(sql)

# Transfer Packages
for i in range(5):
    tx_id = generate_tx_id()
    package_id = random.choice(packages)
    block_number = random.randint(301, 400)
    block_hash = generate_block_hash()
    timestamp = get_timestamp(0)
    
    args = f'["{package_id}", "Distributor A", "Retailer B"]'
    
    sql = f"""INSERT INTO transactions (id, tx_id, channel_name, chaincode_name, function_name, args, status, block_number, block_hash, timestamp) 
    VALUES (gen_random_uuid(), '{tx_id}', 'ibnchannel', 'teatrace', 'TransferPackage', '{args}', 'VALID', {block_number}, '{block_hash}', '{timestamp}');"""
    sql_statements.append(sql)

print("\n".join(sql_statements))
