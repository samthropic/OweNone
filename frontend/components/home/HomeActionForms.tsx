"use client";

import { useActionState, useState, type ReactNode } from "react";
import { useFormStatus } from "react-dom";
import { useRouter } from "next/navigation";
import {
  addFriendAction,
  createGroupAction,
  createSettlementAction,
} from "@/app/actions";
import { AddExpenseDialog } from "@/components/expenses/AddExpenseDialog";
import { initialActionState, type ActionState } from "@/lib/action-state";
import { DEFAULT_CURRENCY } from "@/lib/currencies";
import { useRefreshOnSuccess } from "@/lib/use-refresh-on-success";
import type { Dashboard } from "@/lib/api-types";

export function HeaderActions({ dashboard }: { dashboard: Dashboard }) {
  const router = useRouter();
  const [expenseOpen, setExpenseOpen] = useState(false);
  const [paymentOpen, setPaymentOpen] = useState(false);
  const [paymentState, paymentAction] = useActionState(createSettlementAction, initialActionState);
  useRefreshOnSuccess(paymentState);
  const currency = DEFAULT_CURRENCY;

  return (
    <>
      <div className="home-header-actions">
        <button type="button" className="home-btn" onClick={() => setExpenseOpen(true)} disabled={!dashboard.groups.length}>
          Add expense
        </button>
        <button type="button" className="home-btn home-btn-primary" onClick={() => setPaymentOpen(true)} disabled={!dashboard.friends.length}>
          Pay someone
        </button>
      </div>

      {expenseOpen ? (
        <AddExpenseDialog
          groups={dashboard.groups}
          currentUserId={dashboard.user.id}
          onClose={() => setExpenseOpen(false)}
          onSuccess={() => {
            setExpenseOpen(false);
            router.refresh();
          }}
        />
      ) : null}

      {paymentOpen ? (
        <ActionDialog title="Record a payment" description="Use this after money has been sent." onClose={() => setPaymentOpen(false)}>
          <form action={paymentAction} className="home-action-form">
            <label>
              Recipient
              <select name="toUserId" required>
                {dashboard.friends.map((friend) => <option key={friend.user.id} value={friend.user.id}>{friend.user.displayName}</option>)}
              </select>
            </label>
            <label>
              Amount
              <input name="amount" inputMode="decimal" pattern="\d+(\.\d{1,2})?" required placeholder="0.00" />
            </label>
            <input type="hidden" name="currency" value={currency} />
            <ActionMessage state={paymentState} />
            <SubmitButton label="Record payment" />
          </form>
        </ActionDialog>
      ) : null}
    </>
  );
}

export function AddFriendButton() {
  const [open, setOpen] = useState(false);
  const [state, action] = useActionState(addFriendAction, initialActionState);
  return (
    <>
      <button type="button" className="home-btn home-btn-small" onClick={() => setOpen(true)}>Add friend</button>
      {open ? (
        <ActionDialog title="Add friend" description="They need an existing OweNone account." onClose={() => setOpen(false)}>
          <form action={action} className="home-action-form">
            <label>Email<input name="email" type="email" autoComplete="email" required placeholder="friend@example.com" /></label>
            <ActionMessage state={state} />
            <SubmitButton label="Add friend" />
          </form>
        </ActionDialog>
      ) : null}
    </>
  );
}

export function NewGroupButton({ dashboard }: { dashboard: Dashboard }) {
  const [open, setOpen] = useState(false);
  const [state, action] = useActionState(createGroupAction, initialActionState);
  return (
    <>
      <button type="button" className="home-btn home-btn-small" onClick={() => setOpen(true)}>New group</button>
      {open ? (
        <ActionDialog title="New group" description="Choose the people who share this ledger." onClose={() => setOpen(false)}>
          <form action={action} className="home-action-form">
            <div className="home-form-grid home-form-grid-icon">
              <label>Icon<input name="icon" maxLength={8} defaultValue="GRP" /></label>
              <label>Name<input name="name" maxLength={100} required placeholder="Weekend away" /></label>
            </div>
            <fieldset>
              <legend>Members</legend>
              <div className="home-checkbox-list">
                {dashboard.friends.map((friend) => (
                  <label key={friend.user.id} className="home-checkbox">
                    <input type="checkbox" name="memberIds" value={friend.user.id} />
                    <span>{friend.user.displayName}</span>
                  </label>
                ))}
              </div>
            </fieldset>
            <ActionMessage state={state} />
            <SubmitButton label="Create group" />
          </form>
        </ActionDialog>
      ) : null}
    </>
  );
}

function ActionDialog({ title, description, children, onClose }: { title: string; description: string; children: ReactNode; onClose: () => void }) {
  return (
    <div className="home-modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className="home-modal" role="dialog" aria-modal="true" aria-labelledby="home-modal-title">
        <div className="home-modal-head">
          <div><h2 id="home-modal-title">{title}</h2><p>{description}</p></div>
          <button type="button" className="home-icon-btn" onClick={onClose} aria-label="Close" title="Close">×</button>
        </div>
        {children}
      </section>
    </div>
  );
}

function SubmitButton({ label, className = "home-btn home-btn-primary" }: { label: string; className?: string }) {
  const { pending } = useFormStatus();
  return <button type="submit" className={className} disabled={pending}>{pending ? "Working..." : label}</button>;
}

function ActionMessage({ state, compact = false }: { state: ActionState; compact?: boolean }) {
  if (state.status === "idle") return null;
  return <p className={`home-action-message is-${state.status} ${compact ? "is-compact" : ""}`} role="status">{state.message}</p>;
}