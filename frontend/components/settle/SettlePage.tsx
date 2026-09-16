"use client";

import { useActionState, useEffect, useMemo, useState } from "react";
import { useFormStatus } from "react-dom";
import { createSettlementAction, sendReminderAction } from "@/app/actions";
import { initialActionState, type ActionState } from "@/lib/action-state";
import { DEFAULT_CURRENCY } from "@/lib/currencies";
import { useRefreshOnSuccess } from "@/lib/use-refresh-on-success";
import {
  formatMoney,
  relativeTime,
  shortName,
} from "@/components/home/homeData";
import { UserAvatar } from "@/components/home/UserAvatar";
import type {
  ActivityFeed,
  ActivityItem,
  Dashboard,
  Money,
  NamedTransfer,
  User,
} from "@/lib/api-types";
import { PaymentAppsLinks } from "@/components/settle/PaymentAppsLinks";
import {
  hasPaymentApps,
  paymentMethodLabel,
  type PaymentAppLink,
} from "@/lib/payment-apps";

type Props = {
  dashboard: Dashboard;
  recentPayments: ActivityFeed;
};

export function SettlePage({ dashboard, recentPayments }: Props) {
  const { user, suggestion, summary, friends, groups, netPositions } = dashboard;

  const myTransfers = suggestion.transfers.filter((t) => t.from.id === user.id);
  const waitingTransfers = suggestion.transfers.filter((t) => t.to.id === user.id);
  const isAllSquare =
    suggestion.transfers.length === 0 && summary.netBalance.amountMinor === 0;

  // Unique people the current user can pay — netPositions first (they have a
  // balance), then remaining confirmed friends not already listed
  const customPayPeople = useMemo(() => {
    const seen = new Set<string>();
    const people: User[] = [];
    for (const np of netPositions) {
      if (!seen.has(np.user.id)) {
        seen.add(np.user.id);
        people.push(np.user);
      }
    }
    for (const f of friends) {
      if (f.isFriend && !seen.has(f.user.id)) {
        seen.add(f.user.id);
        people.push(f.user);
      }
    }
    return people;
  }, [netPositions, friends]);

  const currency = DEFAULT_CURRENCY;

  return (
    <>
      <SettleHead suggestion={suggestion} />

      {isAllSquare ? (
        <>
          <AllSquareState />
          <div className="settle-sections">
            <CustomPaymentForm
              people={customPayPeople}
              groups={groups}
              currency={currency}
            />
            {recentPayments.activity.length > 0 && (
              <RecentPayments feed={recentPayments} userId={user.id} />
            )}
          </div>
        </>
      ) : (
        <div className="settle-sections">
          {myTransfers.length > 0 && (
            <section aria-labelledby="settle-pay-heading">
              <SuggestedPaymentsCard transfers={myTransfers} />
            </section>
          )}

          {waitingTransfers.length > 0 && (
            <section
              className="home-card"
              aria-labelledby="settle-waiting-heading"
            >
              <div className="home-card-head">
                <div>
                  <h2 id="settle-waiting-heading">Waiting on others</h2>
                  <p>
                    {waitingTransfers.length}{" "}
                    {waitingTransfers.length === 1 ? "person owes" : "people owe"}{" "}
                    you.
                  </p>
                </div>
              </div>
              <div className="home-list">
                {waitingTransfers.map((transfer) => (
                  <ReminderRow
                    key={`${transfer.from.id}-${transfer.to.id}`}
                    transfer={transfer}
                  />
                ))}
              </div>
            </section>
          )}

          <CustomPaymentForm
            people={customPayPeople}
            groups={groups}
            currency={currency}
          />

          {recentPayments.activity.length > 0 && (
            <RecentPayments feed={recentPayments} userId={user.id} />
          )}
        </div>
      )}
    </>
  );
}

// ————————————————————————————————————————————————
// Header — compression headline
// ————————————————————————————————————————————————

function SettleHead({
  suggestion,
}: {
  suggestion: Dashboard["suggestion"];
}) {
  const hasCompression =
    suggestion.originalPaymentCount > 0 ||
    suggestion.offsettingDebt.amountMinor > 0;

  return (
    <div className="settle-page-head">
      <div>
        <h1>Settle up</h1>
        <p>The minimum transfers to settle every balance in your network.</p>
      </div>
      {hasCompression && (
        <div className="settle-headline" aria-label="Compression summary">
          <span className="settle-stat-chip">
            {suggestion.originalPaymentCount} → {suggestion.reducedPaymentCount} transfers
          </span>
          <span className="settle-stat-chip">
            {formatMoney(suggestion.offsettingDebt, false)} cancelled
          </span>
          {suggestion.groups.length > 0 && (
            <span className="settle-stat-chip">
              {suggestion.groups.length}{" "}
              {suggestion.groups.length === 1 ? "group" : "groups"}
            </span>
          )}
        </div>
      )}
    </div>
  );
}

// ————————————————————————————————————————————————
// Payments to make — pick a transfer, confirm what you're sending
// ————————————————————————————————————————————————

function SuggestedPaymentsCard({ transfers }: { transfers: NamedTransfer[] }) {
  const [selectedKey, setSelectedKey] = useState(
    () => transferKey(transfers[0]),
  );
  const [pendingMethod, setPendingMethod] = useState<
    PaymentAppLink["id"] | "other" | null
  >(null);
  const [state, action] = useActionState(
    createSettlementAction,
    initialActionState,
  );
  useRefreshOnSuccess(state);

  // After confirm, clear the pay/confirm step so the card is ready for the next transfer.
  useEffect(() => {
    if (state.status !== "success") return;
    setPendingMethod(null);
  }, [state]);

  // Keep selection valid when the paid transfer drops out of the list.
  useEffect(() => {
    if (transfers.length === 0) return;
    if (!transfers.some((t) => transferKey(t) === selectedKey)) {
      setSelectedKey(transferKey(transfers[0]));
    }
  }, [transfers, selectedKey]);

  const selected =
    transfers.find((t) => transferKey(t) === selectedKey) ?? transfers[0];
  const hasApps = hasPaymentApps(selected.to.paymentApps);

  function selectTransfer(toUserId: string) {
    const next = transfers.find((t) => t.to.id === toUserId);
    if (next) {
      setSelectedKey(transferKey(next));
      setPendingMethod(null);
    }
  }

  return (
    <div className="settle-pay-card">
      <div className="settle-pay-card-head">
        <h2 id="settle-pay-heading">Payments to make</h2>
        <p>
          {transfers.length}{" "}
          {transfers.length === 1 ? "transfer clears" : "transfers clear"} your
          debts. Open a payment app, send the money, then confirm here.
        </p>
      </div>

      <div className="settle-confirm-form">
        <label className="settle-confirm-field">
          Who are you paying?
          <select
            value={selected.to.id}
            onChange={(e) => selectTransfer(e.target.value)}
            required
          >
            {transfers.map((transfer) => (
              <option key={transferKey(transfer)} value={transfer.to.id}>
                {transfer.to.displayName} —{" "}
                {formatMoney(transfer.money, false)}
              </option>
            ))}
          </select>
        </label>

        <SendingSummary recipient={selected.to} amount={selected.money} />

        {hasApps ? (
          <div className="settle-pay-methods">
            <p className="settle-pay-methods-label">1. Pay with</p>
            <PaymentAppsLinks
              apps={selected.to.paymentApps}
              amountMinor={selected.money.amountMinor}
              compact
              onChoose={(link) => setPendingMethod(link.id)}
              selectedId={
                pendingMethod && pendingMethod !== "other"
                  ? pendingMethod
                  : null
              }
            />
          </div>
        ) : (
          <p className="settle-pay-methods-empty">
            {selected.to.displayName} hasn&apos;t linked Venmo, PayPal, Cash App,
            or Zelle yet. Confirm below after you pay them another way.
          </p>
        )}

        {pendingMethod ? (
          <ConfirmAfterPay
            action={action}
            recipient={selected.to}
            amount={selected.money}
            paymentMethod={pendingMethod}
            state={state}
            onCancel={() => setPendingMethod(null)}
          />
        ) : (
          <div className="settle-confirm-actions">
            {state.status === "success" && (
              <p className="settle-pay-status is-success" role="status">
                {state.message}
              </p>
            )}
            <button
              type="button"
              className="settle-pay-btn settle-pay-btn-secondary"
              onClick={() => setPendingMethod("other")}
            >
              I paid another way
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function ConfirmAfterPay({
  action,
  recipient,
  amount,
  paymentMethod,
  groupId,
  state,
  onCancel,
}: {
  action: (payload: FormData) => void;
  recipient: User;
  amount: Money;
  paymentMethod: PaymentAppLink["id"] | "other";
  groupId?: string;
  state: ActionState;
  onCancel: () => void;
}) {
  const methodName = paymentMethodLabel(paymentMethod);
  const viaApp = paymentMethod !== "other";

  return (
    <form action={action} className="settle-after-pay">
      <p className="settle-after-pay-kicker">2. Confirm in OweNone</p>
      <p className="settle-after-pay-title">
        {viaApp
          ? `Did you send ${formatMoney(amount, false)} to ${recipient.displayName} on ${methodName}?`
          : `Did you send ${formatMoney(amount, false)} to ${recipient.displayName}?`}
      </p>
      <p className="settle-after-pay-sub">
        Payment apps don&apos;t notify OweNone automatically — tap confirm after
        the money is sent.
      </p>
      <input type="hidden" name="toUserId" value={recipient.id} />
      <input
        type="hidden"
        name="amount"
        value={(amount.amountMinor / 100).toFixed(2)}
      />
      <input type="hidden" name="currency" value={amount.currency} />
      <input type="hidden" name="paymentMethod" value={paymentMethod} />
      {groupId ? <input type="hidden" name="groupId" value={groupId} /> : null}
      <div className="settle-after-pay-actions">
        <PaySubmitButton
          label={
            viaApp
              ? `Yes, mark as paid via ${methodName}`
              : "Yes, mark as paid"
          }
        />
        <button
          type="button"
          className="settle-pay-btn settle-pay-btn-secondary"
          onClick={onCancel}
        >
          Not yet
        </button>
      </div>
      <div
        className="settle-pay-status-region"
        role="status"
        aria-live="polite"
      >
        {state.status !== "idle" && (
          <p className={`settle-pay-status is-${state.status}`}>
            {state.message}
          </p>
        )}
      </div>
    </form>
  );
}

function transferKey(transfer: NamedTransfer) {
  return `${transfer.from.id}-${transfer.to.id}`;
}

function SendingSummary({
  recipient,
  amount,
}: {
  recipient: User;
  amount: Money;
}) {
  return (
    <div className="settle-sending" aria-live="polite">
      <UserAvatar user={recipient} className="settle-sending-avatar" />
      <div className="settle-sending-copy">
        <p className="settle-sending-label">You are sending</p>
        <p className="settle-sending-line">
          <span className="settle-sending-amount">
            {formatMoney(amount, false)}
          </span>
          <span className="settle-sending-to">to</span>
          <span className="settle-sending-name">{recipient.displayName}</span>
        </p>
      </div>
    </div>
  );
}

function PaySubmitButton({ label }: { label: string }) {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="settle-pay-btn" disabled={pending}>
      {pending ? "Working…" : label}
    </button>
  );
}

// ————————————————————————————————————————————————
// Waiting on others — each row owns its own action state
// ————————————————————————————————————————————————

function ReminderRow({ transfer }: { transfer: NamedTransfer }) {
  const [state, action] = useActionState(sendReminderAction, initialActionState);

  return (
    <article className="home-row">
      <UserAvatar user={transfer.from} tone="amber" />
      <div className="home-row-copy">
        <p className="home-row-title">{transfer.from.displayName}</p>
        <p className="home-row-sub">
          Owes you {formatMoney(transfer.money, false)}
        </p>
      </div>
      <div className="home-row-right">
        <form action={action} className="home-inline-action">
          <input type="hidden" name="recipientId" value={transfer.from.id} />
          <RemindSubmitButton />
          <div
            className="settle-status-region"
            role="status"
            aria-live="polite"
          >
            {state.status !== "idle" && (
              <p className={`home-action-message is-${state.status} is-compact`}>
                {state.message}
              </p>
            )}
          </div>
        </form>
      </div>
    </article>
  );
}

function RemindSubmitButton() {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="home-pill-btn" disabled={pending}>
      {pending ? "Working…" : "Remind"}
    </button>
  );
}

// ————————————————————————————————————————————————
// Settle a different amount
// ————————————————————————————————————————————————

function CustomPaymentForm({
  people,
  groups,
  currency,
}: {
  people: User[];
  groups: Dashboard["groups"];
  currency: string;
}) {
  const [state, action] = useActionState(createSettlementAction, initialActionState);
  useRefreshOnSuccess(state);
  const hasPeople = people.length > 0;
  const [toUserId, setToUserId] = useState(people[0]?.id ?? "");
  const [amountText, setAmountText] = useState("");
  const [groupId, setGroupId] = useState("");
  const [pendingMethod, setPendingMethod] = useState<
    PaymentAppLink["id"] | "other" | null
  >(null);

  useEffect(() => {
    if (state.status !== "success") return;
    setPendingMethod(null);
    setAmountText("");
    setGroupId("");
  }, [state]);

  const selectedPerson =
    people.find((person) => person.id === toUserId) ?? people[0];
  const amountMinor = parseAmountToMinor(amountText);
  const previewAmount: Money | null =
    selectedPerson && amountMinor != null && amountMinor > 0
      ? { amountMinor, currency }
      : null;
  const hasApps = hasPaymentApps(selectedPerson?.paymentApps);
  const canConfirm = Boolean(selectedPerson && previewAmount);

  const recipientGroups = useMemo(() => {
    if (!selectedPerson) return [];
    return groups.filter((group) =>
      group.members.some((member) => member.id === selectedPerson.id),
    );
  }, [groups, selectedPerson]);

  useEffect(() => {
    if (groupId && !recipientGroups.some((group) => group.id === groupId)) {
      setGroupId("");
    }
  }, [groupId, recipientGroups]);

  return (
    <section className="home-card" aria-labelledby="settle-custom-heading">
      <div className="home-card-head">
        <div>
          <h2 id="settle-custom-heading">Settle a different amount</h2>
          <p>Pay outside OweNone, then confirm here to update balances.</p>
        </div>
      </div>
      <div className="settle-custom-form-wrap">
        {!hasPeople ? (
          <p className="home-muted settle-no-people">
            Add friends to record payments with them.
          </p>
        ) : (
          <div className="home-action-form">
            <div className="home-form-grid">
              <label>
                Who are you paying?
                <select
                  required
                  value={toUserId}
                  onChange={(e) => {
                    setToUserId(e.target.value);
                    setPendingMethod(null);
                    setGroupId("");
                  }}
                >
                  {people.map((person) => (
                    <option key={person.id} value={person.id}>
                      {person.displayName}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Amount
                <input
                  inputMode="decimal"
                  pattern="\d+(\.\d{1,2})?"
                  required
                  placeholder="0.00"
                  value={amountText}
                  onChange={(e) => {
                    setAmountText(e.target.value);
                    setPendingMethod(null);
                  }}
                />
              </label>
            </div>
            <label>
              Group{" "}
              <span className="settle-optional">(optional)</span>
              <select
                value={groupId}
                onChange={(e) => setGroupId(e.target.value)}
              >
                <option value="">No specific group</option>
                {recipientGroups.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name}
                  </option>
                ))}
              </select>
            </label>

            {selectedPerson && previewAmount ? (
              <SendingSummary
                recipient={selectedPerson}
                amount={previewAmount}
              />
            ) : (
              <p className="settle-sending-placeholder">
                Pick a person and amount to preview the payment.
              </p>
            )}

            {canConfirm && hasApps ? (
              <div className="settle-pay-methods">
                <p className="settle-pay-methods-label">1. Pay with</p>
                <PaymentAppsLinks
                  apps={selectedPerson.paymentApps}
                  amountMinor={previewAmount!.amountMinor}
                  compact
                  onChoose={(link) => setPendingMethod(link.id)}
                  selectedId={
                    pendingMethod && pendingMethod !== "other"
                      ? pendingMethod
                      : null
                  }
                />
              </div>
            ) : null}

            {canConfirm && pendingMethod ? (
              <ConfirmAfterPay
                action={action}
                recipient={selectedPerson!}
                amount={previewAmount!}
                paymentMethod={pendingMethod}
                groupId={groupId || undefined}
                state={state}
                onCancel={() => setPendingMethod(null)}
              />
            ) : canConfirm ? (
              <div className="settle-confirm-actions">
                <button
                  type="button"
                  className="home-btn home-btn-primary"
                  onClick={() => setPendingMethod("other")}
                >
                  I paid — confirm next
                </button>
              </div>
            ) : null}

            {state.status !== "idle" && !pendingMethod && (
              <div
                className="settle-status-region"
                role="status"
                aria-live="polite"
              >
                <p className={`home-action-message is-${state.status}`}>
                  {state.message}
                </p>
              </div>
            )}
          </div>
        )}
      </div>
    </section>
  );
}

function parseAmountToMinor(value: string): number | null {
  const trimmed = value.trim();
  if (!trimmed || !/^\d+(\.\d{1,2})?$/.test(trimmed)) return null;
  const dollars = Number(trimmed);
  if (!Number.isFinite(dollars) || dollars <= 0) return null;
  return Math.round(dollars * 100);
}

// ————————————————————————————————————————————————
// Recent payments
// ————————————————————————————————————————————————

function RecentPayments({
  feed,
  userId,
}: {
  feed: ActivityFeed;
  userId: string;
}) {
  return (
    <section className="home-card" aria-labelledby="settle-recent-heading">
      <div className="home-card-head">
        <div>
          <h2 id="settle-recent-heading">Recent payments</h2>
          <p>
            {feed.total} {feed.total === 1 ? "settlement" : "settlements"}
            {feed.hasMore ? "+" : ""}
          </p>
        </div>
      </div>
      <div className="home-list">
        {feed.activity.map((item) => (
          <RecentPaymentRow key={item.id} item={item} userId={userId} />
        ))}
      </div>
    </section>
  );
}

function RecentPaymentRow({
  item,
  userId,
}: {
  item: ActivityItem;
  userId: string;
}) {
  const isOwn = item.actor.id === userId;
  const actorName = isOwn ? "You" : shortName(item.actor);

  return (
    <article className="home-row">
      <div
        className="home-icon settle-settlement-icon"
        aria-hidden="true"
      >
        ✓
      </div>
      <div className="home-row-copy">
        <p className="home-row-title">
          {actorName} settled &ldquo;{item.description || "Payment"}&rdquo;
        </p>
        <p className="home-row-sub">{item.groupName ?? "Cross-group"}</p>
      </div>
      <div className="home-row-right">
        <p
          className={`home-row-amount ${
            item.impact.amountMinor >= 0 ? "is-positive" : "is-negative"
          }`}
        >
          {formatMoney(item.impact)}
        </p>
        <p className="home-row-sub" suppressHydrationWarning>
          {relativeTime(item.occurredAt)}
        </p>
      </div>
    </article>
  );
}

// ————————————————————————————————————————————————
// All-square empty state
// ————————————————————————————————————————————————

function AllSquareState() {
  return (
    <div className="settle-all-square">
      <p className="settle-all-square-icon" aria-hidden="true">
        ✓
      </p>
      <h2 className="settle-all-square-title">Everything is settled.</h2>
      <p className="settle-all-square-sub">
        No balances to clear. All your shared costs are even.
      </p>
    </div>
  );
}
