"use client";

import { useActionState, useState, type ReactNode } from "react";
import { useFormStatus } from "react-dom";
import Link from "next/link";
import {
  createGroupAction,
  updateGroupAction,
  deleteGroupAction,
} from "@/app/actions";
import { initialActionState, type ActionState } from "@/lib/action-state";
import {
  balanceState,
  formatMoney,
  shortName,
} from "@/components/home/homeData";
import type { Dashboard } from "@/lib/api-types";

type Group = Dashboard["groups"][number];

export function GroupsPage({ dashboard }: { dashboard: Dashboard }) {
  const [createOpen, setCreateOpen] = useState(false);
  const isEmpty = dashboard.groups.length === 0;

  return (
    <>
      <div className="groups-page-head">
        <div>
          <h1>Groups</h1>
          <p>Shared ledgers — split expenses and track who owes what.</p>
        </div>
        <button
          type="button"
          className="home-btn home-btn-small"
          onClick={() => setCreateOpen(true)}
          aria-label="Create a new group"
        >
          New group
        </button>
      </div>

      {isEmpty ? (
        <div className="groups-empty" role="region" aria-label="No groups yet">
          <p className="groups-empty-title">No groups yet</p>
          <p className="groups-empty-sub">
            Create a group to start splitting expenses with friends.
          </p>
          <button
            type="button"
            className="home-btn home-btn-primary"
            onClick={() => setCreateOpen(true)}
          >
            New group
          </button>
        </div>
      ) : (
        <section className="home-card" aria-label="Your groups">
          <div className="home-card-head">
            <div>
              <h2>Your groups</h2>
              <p>
                {dashboard.groups.length}{" "}
                {dashboard.groups.length === 1 ? "group" : "groups"}
              </p>
            </div>
          </div>
          <div className="home-list">
            {dashboard.groups.map((group) => (
              <GroupRow key={group.id} group={group} dashboard={dashboard} />
            ))}
          </div>
        </section>
      )}

      {createOpen && (
        <NewGroupDialog
          dashboard={dashboard}
          onClose={() => setCreateOpen(false)}
        />
      )}
    </>
  );
}

function GroupRow({
  group,
  dashboard,
}: {
  group: Group;
  dashboard: Dashboard;
}) {
  const state = balanceState(group.balance.amountMinor);
  const displayNames = group.members.slice(0, 4).map(shortName).join(", ");
  const overflow = group.members.length > 4 ? group.members.length - 4 : 0;

  return (
    <article className="home-row">
      <Link
        href={`/groups/${group.id}`}
        aria-label={`Open ${group.name}`}
        className="groups-row-link"
      >
        <div className="home-icon" aria-hidden="true">
          {group.icon}
        </div>
        <div className="home-row-copy">
          <p className="home-row-title">{group.name}</p>
          <p className="home-row-sub">
            {displayNames}
            {overflow > 0 ? ` +${overflow} more` : ""}
            {" · "}
            {group.members.length}{" "}
            {group.members.length === 1 ? "member" : "members"}
          </p>
        </div>
      </Link>
      <div className="home-row-right">
        <p
          className={`home-row-amount ${
            state.tone === "green"
              ? "is-positive"
              : state.tone === "rose"
                ? "is-negative"
                : ""
          }`}
        >
          {formatMoney(group.balance)}
        </p>
        <p className="home-row-sub">{state.direction}</p>
        {group.isOwner && (
          <div className="groups-row-actions">
            <EditGroupButton group={group} dashboard={dashboard} />
            <DeleteGroupButton group={group} />
          </div>
        )}
      </div>
    </article>
  );
}

function EditGroupButton({
  group,
  dashboard,
}: {
  group: Group;
  dashboard: Dashboard;
}) {
  const [open, setOpen] = useState(false);
  const [state, action] = useActionState(updateGroupAction, initialActionState);

  return (
    <>
      <button
        type="button"
        className="groups-edit-btn"
        onClick={() => setOpen(true)}
        aria-label={`Edit ${group.name}`}
      >
        Edit
      </button>
      {open && (
        <EditGroupDialog
          group={group}
          dashboard={dashboard}
          state={state}
          formAction={action}
          onClose={() => setOpen(false)}
        />
      )}
    </>
  );
}

function DeleteGroupButton({ group }: { group: Group }) {
  const [open, setOpen] = useState(false);
  const [state, action] = useActionState(deleteGroupAction, initialActionState);

  return (
    <>
      <button
        type="button"
        className="groups-delete-btn"
        onClick={() => setOpen(true)}
        aria-label={`Delete ${group.name}`}
      >
        Delete
      </button>
      {open && (
        <DeleteGroupDialog
          group={group}
          state={state}
          formAction={action}
          onClose={() => setOpen(false)}
        />
      )}
    </>
  );
}

function NewGroupDialog({
  dashboard,
  onClose,
}: {
  dashboard: Dashboard;
  onClose: () => void;
}) {
  const [state, action] = useActionState(createGroupAction, initialActionState);
  const trueFriends = dashboard.friends.filter((f) => f.isFriend);

  return (
    <GroupsModal
      title="New group"
      description="Choose the people who share this ledger."
      onClose={onClose}
    >
      <form action={action} className="home-action-form">
        <div className="home-form-grid home-form-grid-icon">
          <label>
            Icon
            <input name="icon" maxLength={8} defaultValue="GRP" />
          </label>
          <label>
            Name
            <input
              name="name"
              maxLength={100}
              required
              placeholder="Weekend away"
            />
          </label>
        </div>
        <fieldset>
          <legend>Members</legend>
          {trueFriends.length === 0 ? (
            <p className="home-muted groups-no-friends">
              Add friends first — only existing friends can join a group.
            </p>
          ) : (
            <div className="home-checkbox-list">
              {trueFriends.map((friend) => (
                <label key={friend.user.id} className="home-checkbox">
                  <input
                    type="checkbox"
                    name="memberIds"
                    value={friend.user.id}
                  />
                  <span>{shortName(friend.user)}</span>
                </label>
              ))}
            </div>
          )}
        </fieldset>
        <div className="groups-status-region" role="status" aria-live="polite">
          {state.status !== "idle" && (
            <p className={`home-action-message is-${state.status}`}>
              {state.message}
            </p>
          )}
        </div>
        <PrimarySubmitButton label="Create group" />
      </form>
    </GroupsModal>
  );
}

function EditGroupDialog({
  group,
  dashboard,
  state,
  formAction,
  onClose,
}: {
  group: Group;
  dashboard: Dashboard;
  state: ActionState;
  formAction: (payload: FormData) => void;
  onClose: () => void;
}) {
  const trueFriends = dashboard.friends.filter((f) => f.isFriend);
  const currentMemberIds = new Set(group.members.map((m) => m.id));
  // Current members are listed even when they are no longer a friend, otherwise
  // saving the form would silently drop them from the group.
  const memberOptions = [
    ...group.members.filter((m) => m.id !== dashboard.user.id),
    ...trueFriends
      .map((f) => f.user)
      .filter((u) => !currentMemberIds.has(u.id) && u.id !== dashboard.user.id),
  ];

  return (
    <GroupsModal
      title="Edit group"
      description={`Update "${group.name}".`}
      onClose={onClose}
    >
      <form action={formAction} className="home-action-form">
        <input type="hidden" name="groupId" value={group.id} />
        <div className="home-form-grid home-form-grid-icon">
          <label>
            Icon
            <input name="icon" maxLength={8} defaultValue={group.icon} />
          </label>
          <label>
            Name
            <input
              name="name"
              maxLength={100}
              required
              defaultValue={group.name}
            />
          </label>
        </div>
        <fieldset>
          <legend>Members</legend>
          <div className="home-checkbox-list">
            <div className="groups-owner-line">
              <span>{shortName(dashboard.user)}</span>
              <span className="groups-owner-badge">Owner</span>
            </div>
            {memberOptions.map((member) => (
              <label key={member.id} className="home-checkbox">
                <input
                  type="checkbox"
                  name="memberIds"
                  value={member.id}
                  defaultChecked={currentMemberIds.has(member.id)}
                />
                <span>{shortName(member)}</span>
              </label>
            ))}
          </div>
          {memberOptions.length === 0 && (
            <p className="home-muted groups-no-friends">
              No friends to add as members.
            </p>
          )}
        </fieldset>
        <div className="groups-status-region" role="status" aria-live="polite">
          {state.status !== "idle" && (
            <p className={`home-action-message is-${state.status}`}>
              {state.message}
            </p>
          )}
        </div>
        <PrimarySubmitButton label="Save changes" />
      </form>
    </GroupsModal>
  );
}

function DeleteGroupDialog({
  group,
  state,
  formAction,
  onClose,
}: {
  group: Group;
  state: ActionState;
  formAction: (payload: FormData) => void;
  onClose: () => void;
}) {
  return (
    <GroupsModal
      title="Delete group"
      description={`Delete "${group.name}"?`}
      onClose={onClose}
    >
      <p className="groups-dialog-confirm-text">
        This removes the group and its settings. Past expenses and payments are
        kept for record-keeping. Groups with recorded expenses cannot be
        deleted.
      </p>
      <div className="groups-status-region" role="status" aria-live="polite">
        {state.status !== "idle" && (
          <p className={`home-action-message is-${state.status}`}>
            {state.message}
          </p>
        )}
      </div>
      <form action={formAction}>
        <input type="hidden" name="groupId" value={group.id} />
        <div className="groups-dialog-actions">
          <button
            type="button"
            className="groups-cancel-btn"
            onClick={onClose}
          >
            Cancel
          </button>
          <DeleteSubmitButton />
        </div>
      </form>
    </GroupsModal>
  );
}

function GroupsModal({
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
      onMouseDown={(event) =>
        event.target === event.currentTarget && onClose()
      }
    >
      <section
        className="home-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="groups-modal-title"
      >
        <div className="home-modal-head">
          <div>
            <h2 id="groups-modal-title">{title}</h2>
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
    <button
      type="submit"
      className="home-btn home-btn-primary"
      disabled={pending}
    >
      {pending ? "Working…" : label}
    </button>
  );
}

function DeleteSubmitButton() {
  const { pending } = useFormStatus();
  return (
    <button
      type="submit"
      className="groups-delete-confirm-btn"
      disabled={pending}
    >
      {pending ? "Working…" : "Delete group"}
    </button>
  );
}
