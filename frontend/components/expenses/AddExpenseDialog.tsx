"use client";

import { useActionState, useEffect, useMemo, useState } from "react";
import { useFormStatus } from "react-dom";
import { createExpenseAction } from "@/app/actions";
import {
  ReceiptSplitter,
  emptyReceiptLine,
  receiptLinesValid,
  sharesFromReceiptLines,
  type ReceiptLine,
} from "@/components/expenses/ReceiptSplitter";
import { initialActionState } from "@/lib/action-state";
import { DEFAULT_CURRENCY } from "@/lib/currencies";
import { formatMinorAsDecimal } from "@/lib/receipt-parse";
import type { Dashboard, User } from "@/lib/api-types";

type GroupOption = Dashboard["groups"][number];

type Props = {
  groups: GroupOption[];
  currentUserId: string;
  fixedGroupId?: string;
  onClose: () => void;
  onSuccess?: () => void;
};

export function AddExpenseDialog({
  groups,
  currentUserId,
  fixedGroupId,
  onClose,
  onSuccess,
}: Props) {
  const [mode, setMode] = useState<"simple" | "receipt">("simple");
  const [selectedGroupID, setSelectedGroupID] = useState(
    fixedGroupId ?? groups[0]?.id ?? "",
  );
  const [lines, setLines] = useState<ReceiptLine[]>([]);
  const [clientError, setClientError] = useState("");
  const [state, action] = useActionState(createExpenseAction, initialActionState);

  const selectedGroup = groups.find((group) => group.id === selectedGroupID);
  const members = selectedGroup?.members ?? [];

  useEffect(() => {
    if (state.status !== "success" || !onSuccess) return;
    const timer = setTimeout(onSuccess, 1000);
    return () => clearTimeout(timer);
  }, [state.status, onSuccess]);

  const shares = useMemo(() => sharesFromReceiptLines(lines), [lines]);
  const receiptTotal = shares.reduce((sum, share) => sum + share.amountMinor, 0);

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    setClientError("");
    if (mode !== "receipt") return;
    const problem = receiptLinesValid(lines);
    if (problem) {
      event.preventDefault();
      setClientError(problem);
      return;
    }
    if (!shares.length || receiptTotal <= 0) {
      event.preventDefault();
      setClientError("Assign items so at least one person owes a share.");
    }
  }

  return (
    <div
      className="home-modal-backdrop"
      onMouseDown={(event) => event.target === event.currentTarget && onClose()}
    >
      <section
        className="home-modal home-modal-wide"
        role="dialog"
        aria-modal="true"
        aria-labelledby="add-expense-title"
      >
        <div className="home-modal-head">
          <div>
            <h2 id="add-expense-title">Add expense</h2>
            <p>Split equally, or assign grocery items from a receipt.</p>
          </div>
          <button type="button" className="home-icon-btn" onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>

        <div className="expense-mode-tabs" role="tablist" aria-label="Expense type">
          <button
            type="button"
            role="tab"
            aria-selected={mode === "simple"}
            className={`expense-mode-tab${mode === "simple" ? " is-on" : ""}`}
            onClick={() => setMode("simple")}
          >
            Simple split
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={mode === "receipt"}
            className={`expense-mode-tab${mode === "receipt" ? " is-on" : ""}`}
            onClick={() => {
              setMode("receipt");
              if (!lines.length) setLines([emptyReceiptLine(members.map((member) => member.id))]);
            }}
          >
            Receipt items
          </button>
        </div>

        <form action={action} className="home-action-form" onSubmit={handleSubmit}>
          {fixedGroupId ? (
            <input type="hidden" name="groupId" value={fixedGroupId} />
          ) : (
            <label>
              Group
              <select
                name="groupId"
                value={selectedGroupID}
                onChange={(event) => {
                  setSelectedGroupID(event.target.value);
                  setLines([]);
                }}
                required
              >
                {groups.map((group) => (
                  <option key={group.id} value={group.id}>
                    {group.name}
                  </option>
                ))}
              </select>
            </label>
          )}

          <label>
            Description
            <input
              name="description"
              maxLength={200}
              required
              placeholder={
                mode === "receipt" ? "Groceries, Costco run…" : "Dinner, tickets, groceries…"
              }
              defaultValue={mode === "receipt" ? "Groceries" : undefined}
              key={mode}
            />
          </label>

          {mode === "simple" ? (
            <>
              <div className="home-form-grid">
                <label>
                  Amount
                  <input
                    name="amount"
                    inputMode="decimal"
                    pattern="\d+(\.\d{1,2})?"
                    required
                    placeholder="0.00"
                  />
                </label>
                <CategorySelect />
              </div>
              <input type="hidden" name="splitMethod" value="equal" />
              <PaidBySelect members={members} currentUserId={currentUserId} />
              <fieldset key={selectedGroupID}>
                <legend>Split equally between</legend>
                <div className="home-checkbox-list">
                  {members.map((member) => (
                    <label key={member.id} className="home-checkbox">
                      <input
                        type="checkbox"
                        name="participantIds"
                        value={member.id}
                        defaultChecked
                      />
                      <span>{member.displayName}</span>
                    </label>
                  ))}
                </div>
              </fieldset>
            </>
          ) : (
            <>
              <CategorySelect defaultValue="food" />
              <PaidBySelect members={members} currentUserId={currentUserId} />
              <ReceiptSplitter
                key={selectedGroupID}
                members={members}
                lines={lines}
                onChange={setLines}
              />
              <input type="hidden" name="splitMethod" value="exact" />
              <input
                type="hidden"
                name="amount"
                value={formatMinorAsDecimal(receiptTotal || 0)}
              />
              <input type="hidden" name="shares" value={JSON.stringify(shares)} />
              {shares.length ? (
                <ul className="receipt-share-summary">
                  {shares.map((share) => {
                    const member = members.find((entry) => entry.id === share.userId);
                    return (
                      <li key={share.userId}>
                        <span>{member?.displayName ?? "Member"}</span>
                        <span>${formatMinorAsDecimal(share.amountMinor)}</span>
                      </li>
                    );
                  })}
                </ul>
              ) : null}
            </>
          )}

          <input type="hidden" name="currency" value={DEFAULT_CURRENCY} />

          {(clientError || state.status !== "idle") && (
            <p
              className={`home-action-message is-${clientError ? "error" : state.status}`}
              role="status"
            >
              {clientError || state.message}
            </p>
          )}

          <SubmitButton label="Add expense" />
        </form>
      </section>
    </div>
  );
}

function CategorySelect({ defaultValue = "general" }: { defaultValue?: string }) {
  return (
    <label>
      Category
      <select name="category" defaultValue={defaultValue}>
        <option value="general">General</option>
        <option value="food">Food</option>
        <option value="travel">Travel</option>
        <option value="home">Home</option>
      </select>
    </label>
  );
}

function PaidBySelect({
  members,
  currentUserId,
}: {
  members: User[];
  currentUserId: string;
}) {
  return (
    <label>
      Paid by
      <select name="paidByUserId" defaultValue={currentUserId} required>
        {members.map((member) => (
          <option key={member.id} value={member.id}>
            {member.displayName}
          </option>
        ))}
      </select>
    </label>
  );
}

function SubmitButton({ label }: { label: string }) {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="home-btn home-btn-primary" disabled={pending}>
      {pending ? "Working…" : label}
    </button>
  );
}
