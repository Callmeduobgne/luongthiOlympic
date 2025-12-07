export type TeaPackageStatus = "CREATED" | "VERIFIED" | "SOLD" | "EXPIRED";

export interface TeaPackage {
  docType: "package"; // Document type for CouchDB indexing
  packageId: string;
  batchId: string;
  blockHash: string; // Blockhash từ transaction khi tạo package
  hashVersion?: string; // "v1" (no secret) or "v2" (with secret) - for backward compatibility
  txId: string; // Transaction ID khi tạo package
  weight: number; // Trọng lượng gói (gram)
  productionDate: string; // Ngày sản xuất (YYYY-MM-DD)
  expiryDate?: string; // Hạn sử dụng (YYYY-MM-DD, optional)
  qrCode?: string; // QR code data (optional)
  status: TeaPackageStatus;
  owner: string; // MSP ID
  timestamp: string; // ISO timestamp
}

export interface CreateTeaPackageInput {
  packageId: string;
  batchId: string;
  weight: number;
  productionDate: string;
  expiryDate?: string;
  qrCode?: string;
}

export function isTeaPackageStatus(value: string): value is TeaPackageStatus {
  return value === "CREATED" || value === "VERIFIED" || value === "SOLD" || value === "EXPIRED";
}
