"use client";

import { useActionState, useState, type ReactNode } from "react";
import { useFormStatus } from "react-dom";
import { addFriendAction, removeFriendAction } from "@/app/actions";
import { initialActionState, type ActionState } from "@/lib/action-state";
import { balanceState, formatMoney, shortName } from "@/components/home/homeData";
import { UserAvatar } from "@/components/home/UserAvatar";
import type { Dashboard } from "@/lib/api-types";

type FriendEntry = Dashboard["friends"][number];

export function FriendsPage({ dashboard }: { dashboard: Dashboard }) {
  const trueFriends = dashboard.friends.filter((f) => f.isFriend);
  const groupOnly = dashboard.friends.filter((f) => !f.isFriend);
  const isEmpty = dashboard.friends.length === 0;

  const [addOpen, setAddOpen] = useState(false);

  return (
    <>
      <div className="friends-page-head">
        <div>
          <h1>Friends</h1>
          <p>Everyone you share money with</p>
        </div>
        <button
          type="button"
          className="home-btn home-btn-small"
          onClick={() => setAddOpen(true)}
          aria-label="Add a friend by email"
        >
          Add friend
        </button>
      </div>

      {isEmpty ? (
        <div className="friends-empty" role="region" aria-label="No friends yet">
          <p className="friends-empty-title">No connections yet</p>
          <p className="friends-empty-sub">Add a friend to start tracking shared expenses.</p>
          <button
            type="button"
            className="home-btn home-btn-primary"
            onClick={() => setAddOpen(true)}
          >
            Add friend
          </button>
        </div>
      ) : (
        <div className="friends-lists">
          {trueFriends.length > 0 && (
            <section className="home-card" aria-label="Friends">
              <div className="home-card-head">
                <div>
                  <h2>Friends</h2>
                  <p>
                    {trueFriends.length} {trueFriends.length === 1 ? "person" : "people"}
                  </p>
                </div>
              </div>
              <div className="home-list">
                {trueFriends.map((friend) => (
                  <FriendRow key={friend.user.id} friend={friend} isFriend />
                ))}
              </div>
            </section>
          )}

          {groupOnly.length > 0 && (
            <section className="home-card" aria-label="Also in your groups">
              <div className="home-card-head">
                <div>
                  <h2>Also in your groups</h2>
                  <p>Connected through shared groups — not yet a friend</p>
                </div>
              </div>
              <div className="home-list">
                {groupOnly.map((friend) => (
                  <FriendRow key={friend.user.id} friend={friend} isFriend={false} />
                ))}
              </div>
            </section>
          )}
        </div>
      )}

      {addOpen && <AddFriendDialog onClose={() => setAddOpen(false)} />}
    </>
  );
}

function FriendRow({ friend, isFriend }: { friend: FriendEntry; isFriend: boolean }) {
  const state = balanceState(friend.balance.amountMinor);
  return (
    <article className="home-row">
      <UserAvatar user={friend.user} tone={state.tone} />
      <div className="home-row-copy">
        <p className="home-row-title">{shortName(friend.user)}</p>
        <p className="home-row-sub">{friend.groupNames.join(", ") || "No shared groups"}</p>
        {friend.user.email ? (
          <p className="friends-row-email">{friend.user.email}</p>
        ) : null}
      </div>
      <div className="home-row-right">
        <p className="home-row-amount">{formatMoney(friend.balance)}</p>
        <p className="home-row-sub">{state.direction}</p>
        {isFriend ? (
          <RemoveFriendButton friend={friend} />
        ) : (
          <AddGroupContactButton email={friend.user.email ?? ""} />
        )}
      </div>
    </article>
  );
}

function RemoveFriendButton({ friend }: { friend: FriendEntry }) {
  const [open, setOpen] = useState(false);
  const [state, action] = useActionState(removeFriendAction, initialActionState);

  // Dialog closes naturally when the friend is removed: revalidatePath causes a re-render
  // that unmounts this row. If the action errors (e.g. outstanding balance), the dialog
  // stays open so the user can read the API's message.

  return (
    <>
      <button
        type="button"
        className="friends-remove-btn"
        onClick={() => setOpen(true)}
        aria-label={`Remove ${friend.user.displayName} as a friend`}
      >
        Remove
      </button>
      {open && (
        <RemoveFriendDialog
          friend={friend}
          state={state}
          formAction={action}
          onClose={() => setOpen(false)}
        />
      )}
    </>
  );
}

function AddGroupContactButton({ email }: { email: string }) {
  const [state, action] = useActionState(addFriendAction, initialActionState);
  if (!email) return null;
  return (
    <form action={action} className="home-inline-action">
      <input type="hidden" name="email" value={email} />
      <AddSubmitButton label="Add friend" />
      <div className="friends-status-region" role="status" aria-live="polite">
        {state.status !== "idle" && (
          <p className={`home-action-message is-${state.status} is-compact`}>{state.message}</p>
        )}
      </div>
    </form>
  );
}

function AddFriendDialog({ onClose }: { onClose: () => void }) {
  const [state, action] = useActionState(addFriendAction, initialActionState);
  return (
    <FriendsModal
      title="Add friend"
      description="They need an existing OweNone account."
      onClose={onClose}
    >
      <form action={action} className="home-action-form">
        <label>
          Email
          <input
            name="email"
            type="email"
            autoComplete="email"
            required
            placeholder="friend@example.com"
          />
        </label>
        <div className="friends-status-region" role="status" aria-live="polite">
          {state.status !== "idle" && (
            <p className={`home-action-message is-${state.status}`}>{state.message}</p>
          )}
        </div>
        <PrimarySubmitButton label="Add friend" />
      </form>
    </FriendsModal>
  );
}

function RemoveFriendDialog({
  friend,
  state,
  formAction,
  onClose,
}: {
  friend: FriendEntry;
  state: ActionState;
  formAction: (payload: FormData) => void;
  onClose: () => void;
}) {
  return (
    <FriendsModal
      title="Remove friend"
      description={`Remove ${friend.user.displayName} from your friends list?`}
      onClose={onClose}
    >
      <p className="friends-dialog-confirm-text">
        Your shared groups and past expenses stay as they are. You can only remove someone once
        your balance with them is settled.
      </p>
      <div className="friends-status-region" role="status" aria-live="polite">
        {state.status !== "idle" && (
          <p className={`home-action-message is-${state.status}`}>{state.message}</p>
        )}
      </div>
      <form action={formAction}>
        <input type="hidden" name="email" value={friend.user.email ?? ""} />
        <div className="friends-dialog-actions">
          <button type="button" className="friends-cancel-btn" onClick={onClose}>
            Cancel
          </button>
          <RemoveSubmitButton />
        </div>
      </form>
    </FriendsModal>
  );
}

function FriendsModal({
  title,
  description,
  children,
  onClose,
}: {
  title: string;
  description: string;
  children: ReactNode;
  onClose: () => void;
}) {
  return (
    <div
      className="home-modal-backdrop"
      onMouseDown={(event) => event.target === event.currentTarget && onClose()}
    >
      <section
        className="home-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="friends-modal-title"
      >
        <div className="home-modal-head">
          <div>
            <h2 id="friends-modal-title">{title}</h2>
            <p>{description}</p>
          </div>
          <button
            type="button"
            className="home-icon-btn"
            onClick={onClose}
            aria-label="Close"
            title="Close"
          >
            ×
          </button>
        </div>
        {children}
      </section>
    </div>
  );
}

function PrimarySubmitButton({ label }: { label: string }) {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="home-btn home-btn-primary" disabled={pending}>
      {pending ? "Working…" : label}
    </button>
  );
}

function AddSubmitButton({ label }: { label: string }) {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="home-pill-btn" disabled={pending}>
      {pending ? "Working…" : label}
    </button>
  );
}

function RemoveSubmitButton() {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="friends-confirm-btn" disabled={pending}>
      {pending ? "Working…" : "Remove friend"}
    </button>
  );
}
