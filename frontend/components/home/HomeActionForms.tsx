"use client";

import { useActionState, useState, type ReactNode } from "react";
import { useFormStatus } from "react-dom";
import {
  addFriendAction,
  createExpenseAction,
  createGroupAction,
  createSettlementAction,
  initialActionState,
  sendReminderAction,
  type ActionState,
} from "@/app/actions";
import type { Dashboard, NamedTransfer } from "@/lib/api-types";

export function HeaderActions({ dashboard }: { dashboard: Dashboard }) {
  const [expenseOpen, setExpenseOpen] = useState(false);
  const [paymentOpen, setPaymentOpen] = useState(false);
  const [selectedGroupID, setSelectedGroupID] = useState(dashboard.groups[0]?.id ?? "");
  const [expenseState, expenseAction] = useActionState(createExpenseAction, initialActionState);
  const [paymentState, paymentAction] = useActionState(createSettlementAction, initialActionState);
  const selectedGroup = dashboard.groups.find((group) => group.id === selectedGroupID);
  const currency = dashboard.user.preferredCurrency ?? "GBP";

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
        <ActionDialog title="Add expense" description="Balances update across the full network." onClose={() => setExpenseOpen(false)}>
          <form action={expenseAction} className="home-action-form">
            <label>
              Group
              <select name="groupId" value={selectedGroupID} onChange={(event) => setSelectedGroupID(event.target.value)} required>
                {dashboard.groups.map((group) => <option key={group.id} value={group.id}>{group.name}</option>)}
              </select>
            </label>
            <label>
              Description
              <input name="description" maxLength={200} required placeholder="Dinner, tickets, groceries..." />
            </label>
            <div className="home-form-grid">
              <label>
                Amount
                <input name="amount" inputMode="decimal" pattern="\d+(\.\d{1,2})?" required placeholder="0.00" />
              </label>
              <label>
                Category
                <select name="category" defaultValue="general">
                  <option value="general">General</option>
                  <option value="food">Food</option>
                  <option value="travel">Travel</option>
                  <option value="home">Home</option>
                </select>
              </label>
            </div>
            <input type="hidden" name="currency" value={currency} />
            <label>
              Paid by
              <select name="paidByUserId" defaultValue={dashboard.user.id} required>
                {selectedGroup?.members.map((member) => <option key={member.id} value={member.id}>{member.displayName}</option>)}
              </select>
            </label>
            <fieldset key={selectedGroupID}>
              <legend>Split equally between</legend>
              <div className="home-checkbox-list">
                {selectedGroup?.members.map((member) => (
                  <label key={member.id} className="home-checkbox">
                    <input type="checkbox" name="participantIds" value={member.id} defaultChecked />
                    <span>{member.displayName}</span>
                  </label>
                ))}
              </div>
            </fieldset>
            <ActionMessage state={expenseState} />
            <SubmitButton label="Add expense" />
          </form>
        </ActionDialog>
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
              <label>Icon<input name="icon" maxLength={8} defaultValue="👥" /></label>
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

export function SettlementButton({ transfer, label, className }: { transfer: NamedTransfer; label: string; className: string }) {
  const [state, action] = useActionState(createSettlementAction, initialActionState);
  return (
    <form action={action} className="home-inline-action">
      <input type="hidden" name="toUserId" value={transfer.to.id} />
      <input type="hidden" name="amount" value={(transfer.money.amountMinor / 100).toFixed(2)} />
      <input type="hidden" name="currency" value={transfer.money.currency} />
      <SubmitButton label={label} className={className} />
      <ActionMessage state={state} compact />
    </form>
  );
}

export function FriendActionButton({ friend }: { friend: Dashboard["friends"][number] }) {
  const action = friend.balance.amountMinor < 0 ? createSettlementAction : sendReminderAction;
  const [state, formAction] = useActionState(action, initialActionState);
  const isPayment = friend.balance.amountMinor < 0;
  return (
    <form action={formAction} className="home-inline-action">
      {isPayment ? (
        <>
          <input type="hidden" name="toUserId" value={friend.user.id} />
          <input type="hidden" name="amount" value={(Math.abs(friend.balance.amountMinor) / 100).toFixed(2)} />
          <input type="hidden" name="currency" value={friend.balance.currency} />
        </>
      ) : (
        <input type="hidden" name="recipientId" value={friend.user.id} />
      )}
      <SubmitButton label={isPayment ? "Pay" : "Remind"} className="home-pill-btn" />
      <ActionMessage state={state} compact />
    </form>
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