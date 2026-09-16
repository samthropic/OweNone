"use client";

import { useMemo, useState } from "react";
import {
  formatMinorAsDecimal,
  parseDecimalToMinor,
  parseReceiptText,
  splitAmountAmong,
} from "@/lib/receipt-parse";
import type { User } from "@/lib/api-types";

export type ReceiptLine = {
  id: string;
  name: string;
  amountMinor: number;
  assigneeIds: string[];
};

export { formatMinorAsDecimal };

type Props = {
  members: User[];
  lines: ReceiptLine[];
  onChange: (lines: ReceiptLine[]) => void;
};

function newLineId() {
  return crypto.randomUUID();
}

export function emptyReceiptLine(memberIds: string[] = []): ReceiptLine {
  return {
    id: newLineId(),
    name: "",
    amountMinor: 0,
    assigneeIds: [...memberIds],
  };
}

export function ReceiptSplitter({ members, lines, onChange }: Props) {
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [ocrStatus, setOcrStatus] = useState<"idle" | "working" | "done" | "error">("idle");
  const [ocrMessage, setOcrMessage] = useState("");

  const memberIds = members.map((member) => member.id);
  const everyoneSelected = (assignees: string[]) =>
    memberIds.length > 0 && memberIds.every((id) => assignees.includes(id));

  function updateLine(id: string, patch: Partial<ReceiptLine>) {
    onChange(lines.map((line) => (line.id === id ? { ...line, ...patch } : line)));
  }

  function applyParsed(parsed: ReturnType<typeof parseReceiptText>) {
    const { items, receiptTotalMinor } = parsed;
    if (!items.length) {
      setOcrMessage("No priced items found. Try a clearer photo or add items manually.");
      setOcrStatus("error");
      return;
    }
    onChange(
      items.map((item) => ({
        id: newLineId(),
        name: item.name,
        amountMinor: item.amountMinor,
        assigneeIds: [...memberIds],
      })),
    );
    setOcrStatus("done");
    const total = receiptTotalMinor ?? items.reduce((sum, item) => sum + item.amountMinor, 0);
    setOcrMessage(
      `Parsed ${items.length} item${items.length === 1 ? "" : "s"} · total $${formatMinorAsDecimal(total)}.`,
    );
  }

  async function handleImage(file: File) {
    setOcrStatus("working");
    setOcrMessage("Reading receipt…");
    const objectUrl = URL.createObjectURL(file);
    setPreviewUrl((previous) => {
      if (previous) URL.revokeObjectURL(previous);
      return objectUrl;
    });

    try {
      const { createWorker } = await import("tesseract.js");
      const worker = await createWorker("eng");
      const result = await worker.recognize(file);
      await worker.terminate();
      const parsed = parseReceiptText(result.data.text);
      if (!parsed.items.length) {
        setOcrMessage("Could not read prices from that photo. Add items manually.");
        setOcrStatus("error");
        return;
      }
      applyParsed(parsed);
    } catch {
      setOcrStatus("error");
      setOcrMessage("Could not read that image. Try a clearer photo.");
    }
  }

  function toggleAssignee(line: ReceiptLine, userId: string) {
    if (everyoneSelected(line.assigneeIds)) {
      updateLine(line.id, { assigneeIds: [userId] });
      return;
    }
    const has = line.assigneeIds.includes(userId);
    let next = has
      ? line.assigneeIds.filter((id) => id !== userId)
      : [...line.assigneeIds, userId];
    if (next.length === 0) next = [...memberIds];
    updateLine(line.id, { assigneeIds: next });
  }

  function setEveryone(line: ReceiptLine) {
    updateLine(line.id, { assigneeIds: [...memberIds] });
  }

  const totalMinor = useMemo(
    () => lines.reduce((sum, line) => sum + Math.max(0, line.amountMinor), 0),
    [lines],
  );

  return (
    <div className="receipt-splitter">
      <div className="receipt-upload-row">
        <label className="home-btn receipt-upload-btn">
          Upload receipt
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp,image/gif"
            onChange={(event) => {
              const file = event.currentTarget.files?.[0];
              if (file) void handleImage(file);
              event.currentTarget.value = "";
            }}
          />
        </label>
        {previewUrl ? (
          // eslint-disable-next-line @next/next/no-img-element -- local object URL preview
          <img src={previewUrl} alt="Uploaded receipt preview" className="receipt-preview" />
        ) : null}
        <button
          type="button"
          className="home-btn home-btn-small"
          onClick={() => onChange([emptyReceiptLine(memberIds)])}
        >
          Add item
        </button>
      </div>

      {ocrStatus !== "idle" && ocrMessage ? (
        <p className={`receipt-ocr-msg is-${ocrStatus === "working" ? "idle" : ocrStatus}`} role="status">
          {ocrMessage}
        </p>
      ) : null}

      {lines.length ? (
        <ul className="receipt-lines">
          {lines.map((line) => {
            const isEveryone = everyoneSelected(line.assigneeIds);
            return (
              <li key={line.id} className="receipt-line">
                <div className="receipt-line-fields">
                  <input
                    aria-label="Item name"
                    value={line.name}
                    placeholder="Item"
                    onChange={(event) => updateLine(line.id, { name: event.target.value })}
                  />
                  <input
                    aria-label="Item price"
                    inputMode="decimal"
                    placeholder="0.00"
                    value={line.amountMinor > 0 ? formatMinorAsDecimal(line.amountMinor) : ""}
                    onChange={(event) => {
                      const parsed = parseDecimalToMinor(event.target.value);
                      updateLine(line.id, { amountMinor: parsed ?? 0 });
                    }}
                  />
                  <button
                    type="button"
                    className="home-icon-btn"
                    aria-label="Remove item"
                    onClick={() => onChange(lines.filter((entry) => entry.id !== line.id))}
                  >
                    ×
                  </button>
                </div>
                <div className="receipt-assignees" role="group" aria-label={`Who is ${line.name || "this item"} for`}>
                  <button
                    type="button"
                    className={`receipt-chip${isEveryone ? " is-on" : ""}`}
                    onClick={() => setEveryone(line)}
                  >
                    Everyone
                  </button>
                  {members.map((member) => {
                    const on = !isEveryone && line.assigneeIds.includes(member.id);
                    return (
                      <button
                        key={member.id}
                        type="button"
                        className={`receipt-chip${on ? " is-on" : ""}`}
                        onClick={() => toggleAssignee(line, member.id)}
                      >
                        {member.displayName.split(" ")[0]}
                      </button>
                    );
                  })}
                </div>
              </li>
            );
          })}
        </ul>
      ) : (
        <p className="receipt-empty">Upload a receipt photo to pull items, or add them manually.</p>
      )}

      <p className="receipt-total">
        Receipt total <strong>${formatMinorAsDecimal(totalMinor)}</strong>
      </p>
    </div>
  );
}

/** Roll receipt lines into exact expense shares.
 *  Lines that share the same assignee set are summed first, then split once.
 *  That keeps an all-Everyone receipt equal (within at most n-1 cents), instead of
 *  giving the same person the leftover cent on every item. */
export function sharesFromReceiptLines(lines: ReceiptLine[]): Array<{ userId: string; amountMinor: number }> {
  const buckets = new Map<string, { userIds: string[]; amountMinor: number }>();

  for (const line of lines) {
    if (line.amountMinor <= 0 || !line.assigneeIds.length) continue;
    const userIds = [...new Set(line.assigneeIds)].sort((left, right) => left.localeCompare(right));
    const key = userIds.join("|");
    const existing = buckets.get(key);
    if (existing) {
      existing.amountMinor += line.amountMinor;
    } else {
      buckets.set(key, { userIds, amountMinor: line.amountMinor });
    }
  }

  const totals = new Map<string, number>();
  for (const bucket of buckets.values()) {
    const parts = splitAmountAmong(bucket.amountMinor, bucket.userIds.length);
    bucket.userIds.forEach((userId, index) => {
      totals.set(userId, (totals.get(userId) ?? 0) + parts[index]);
    });
  }

  return [...totals.entries()]
    .filter(([, amountMinor]) => amountMinor > 0)
    .map(([userId, amountMinor]) => ({ userId, amountMinor }))
    .sort((left, right) => left.userId.localeCompare(right.userId));
}

export function receiptLinesValid(lines: ReceiptLine[]) {
  if (!lines.length) return "Add at least one receipt item.";
  if (lines.some((line) => !line.name.trim())) return "Every item needs a name.";
  if (lines.some((line) => line.amountMinor <= 0)) return "Every item needs a price.";
  if (lines.some((line) => line.assigneeIds.length === 0)) {
    return "Assign each item to at least one person (or Everyone).";
  }
  return null;
}
