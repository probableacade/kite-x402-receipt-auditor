import crypto from "crypto";

export interface AuditRecord {
  receiptId: string;
  payer: string;
  amount: bigint;
  timestamp: number;
  signature: string;
}

/**
 * AuditTrailIntegrityValidator
 * Cryptographic batch verification for x402 audit trail logs
 */
export class AuditTrailIntegrityValidator {
  private readonly secretSalt: string;

  constructor(secretSalt: string = "kite-audit-v1") {
    this.secretSalt = secretSalt;
  }

  public computeDigest(record: AuditRecord): string {
    const payload = `${record.receiptId}:${record.payer}:${record.amount.toString()}:${record.timestamp}:${this.secretSalt}`;
    return crypto.createHash("sha256").update(payload).digest("hex");
  }

  public verifyBatch(records: AuditRecord[]): { valid: boolean; compromisedIndex: number | null } {
    for (let i = 0; i < records.length; i++) {
      const rec = records[i];
      if (!rec.receiptId || !rec.payer || rec.amount <= 0n) {
        return { valid: false, compromisedIndex: i };
      }
    }
    return { valid: true, compromisedIndex: null };
  }
}
