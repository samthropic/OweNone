"use client";

import { useActionState, useState } from "react";
import { useFormStatus } from "react-dom";
import { updateProfileAction, changePasswordAction, uploadAvatarAction, removeAvatarAction } from "@/app/actions";
import { LogoutForm } from "@/components/auth/LogoutForm";
import { PaymentAppIcon } from "@/components/settle/PaymentAppIcon";
import { UserAvatar } from "@/components/home/UserAvatar";
import { initialActionState } from "@/lib/action-state";
import { DEFAULT_CURRENCY } from "@/lib/currencies";
import type { AuthUser } from "@/lib/api-types";

export function ProfilePage({ user }: { user: AuthUser }) {
  return (
    <>
      <div className="profile-page-head">
        <h1>Profile</h1>
        <p>Manage your details and keep your account secure.</p>
      </div>
      <div className="profile-cards">
        <IdentityCard user={user} />
        <DetailsCard user={user} />
        <PaymentAppsCard user={user} />
        <PasswordCard />
        <LogoutCard />
      </div>
    </>
  );
}

function IdentityCard({ user }: { user: AuthUser }) {
  const [uploadState, uploadAction] = useActionState(uploadAvatarAction, initialActionState);
  const [removeState, removeAction] = useActionState(removeAvatarAction, initialActionState);
  const status =
    uploadState.status !== "idle"
      ? uploadState
      : removeState.status !== "idle"
        ? removeState
        : null;

  return (
    <section className="home-card" aria-label="Account identity">
      <div className="profile-identity">
        <form action={uploadAction} className="profile-avatar-upload">
          <label className="profile-avatar-hit" title={user.avatarUrl ? "Change photo" : "Add photo"}>
            <UserAvatar user={user} className="profile-avatar-lg" size="lg" />
            <span className="profile-avatar-hint">
              {user.avatarUrl ? "Change" : "Add photo"}
            </span>
            <input
              type="file"
              name="avatar"
              accept="image/jpeg,image/png,image/webp,image/gif"
              aria-label={user.avatarUrl ? "Change profile photo" : "Add profile photo"}
              onChange={(event) => {
                const form = event.currentTarget.form;
                if (form && event.currentTarget.files?.length) {
                  form.requestSubmit();
                }
              }}
            />
          </label>
        </form>
        <div className="profile-identity-copy">
          <p className="profile-identity-name">{user.displayName}</p>
          <p className="profile-identity-email">{user.email}</p>
          {user.avatarUrl ? (
            <form action={removeAction} className="profile-avatar-actions">
              <button type="submit" className="profile-avatar-remove">
                Remove photo
              </button>
            </form>
          ) : null}
          {status ? (
            <p className={`home-action-message is-${status.status}`}>{status.message}</p>
          ) : null}
        </div>
      </div>
    </section>
  );
}

function DetailsCard({ user }: { user: AuthUser }) {
  const [state, formAction] = useActionState(updateProfileAction, initialActionState);

  return (
    <section className="home-card" aria-labelledby="details-card-title">
      <div className="home-card-head">
        <div>
          <h2 id="details-card-title">Your details</h2>
          <p>Update your name and email address.</p>
        </div>
      </div>
      <form action={formAction} className="home-action-form profile-card-form">
        <div className="home-form-grid">
          <label>
            Display name
            <input
              name="displayName"
              type="text"
              defaultValue={user.displayName}
              autoComplete="name"
              required
              minLength={1}
              maxLength={100}
            />
          </label>
          <label>
            Email
            <input
              name="email"
              type="email"
              defaultValue={user.email}
              autoComplete="email"
              required
            />
          </label>
        </div>
        <div className="profile-currency-wrap">
          <p className="home-kicker">Currency</p>
          <p className="profile-currency-hint" id="currency-hint">
            OweNone uses {DEFAULT_CURRENCY} only. All balances and payments are
            recorded in US dollars.
          </p>
        </div>
        <div className="profile-status-region" role="status" aria-live="polite">
          {state.status !== "idle" && (
            <p className={`home-action-message is-${state.status}`}>{state.message}</p>
          )}
        </div>
        <SubmitButton label="Save changes" pendingLabel="Saving…" />
      </form>
    </section>
  );
}

function PaymentAppsCard({ user }: { user: AuthUser }) {
  const [state, formAction] = useActionState(updateProfileAction, initialActionState);
  const apps = user.paymentApps ?? {};

  return (
    <section className="home-card" aria-labelledby="payment-apps-title">
      <div className="home-card-head">
        <div>
          <h2 id="payment-apps-title">Payment apps</h2>
          <p>
            Link your Venmo, PayPal, Cash App, or Zelle so friends can pay you
            faster when they settle up.
          </p>
        </div>
      </div>
      <form action={formAction} className="home-action-form profile-card-form">
        <input type="hidden" name="displayName" value={user.displayName} />
        <input type="hidden" name="email" value={user.email} />
        <div className="home-form-grid">
          <label>
            <span className="profile-pay-label">
              <PaymentAppIcon id="venmo" className="profile-pay-icon" />
              Venmo username
            </span>
            <input
              name="venmo"
              type="text"
              defaultValue={apps.venmo ?? ""}
              placeholder="username"
              maxLength={64}
              autoComplete="off"
            />
          </label>
          <label>
            <span className="profile-pay-label">
              <PaymentAppIcon id="paypal" className="profile-pay-icon" />
              PayPal.me
            </span>
            <input
              name="paypal"
              type="text"
              defaultValue={apps.paypal ?? ""}
              placeholder="username"
              maxLength={64}
              autoComplete="off"
            />
          </label>
          <label>
            <span className="profile-pay-label">
              <PaymentAppIcon id="cashApp" className="profile-pay-icon" />
              Cash App cashtag
            </span>
            <input
              name="cashApp"
              type="text"
              defaultValue={apps.cashApp ?? ""}
              placeholder="$cashtag"
              maxLength={64}
              autoComplete="off"
            />
          </label>
          <label>
            <span className="profile-pay-label">
              <PaymentAppIcon id="zelle" className="profile-pay-icon" />
              Zelle email or phone
            </span>
            <input
              name="zelle"
              type="text"
              defaultValue={apps.zelle ?? ""}
              placeholder="name@email.com"
              maxLength={128}
              autoComplete="off"
            />
          </label>
        </div>
        <p className="profile-currency-hint">
          Leave a field blank to unlink that app. Handles are shown to people
          who owe you when they settle up.
        </p>
        <div className="profile-status-region" role="status" aria-live="polite">
          {state.status !== "idle" && (
            <p className={`home-action-message is-${state.status}`}>{state.message}</p>
          )}
        </div>
        <SubmitButton label="Save payment apps" pendingLabel="Saving…" />
      </form>
    </section>
  );
}

function PasswordCard() {
  const [state, formAction] = useActionState(changePasswordAction, initialActionState);
  const [clientError, setClientError] = useState("");

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    const form = e.currentTarget;
    const newPw = (form.elements.namedItem("newPassword") as HTMLInputElement)?.value ?? "";
    const confirmPw =
      (form.elements.namedItem("confirmPassword") as HTMLInputElement)?.value ?? "";
    if (newPw !== confirmPw) {
      e.preventDefault();
      setClientError("Passwords do not match.");
    } else {
      setClientError("");
    }
  }

  // Clear the client-side mismatch error once the action resolves
  const visibleError = clientError || (state.status !== "idle" ? state.message : "");
  const visibleStatus = clientError ? "error" : state.status;

  return (
    <section className="home-card" aria-labelledby="password-card-title">
      <div className="home-card-head">
        <div>
          <h2 id="password-card-title">Password</h2>
          <p>Keep your account secure with a strong password.</p>
        </div>
      </div>
      <form
        action={formAction}
        onSubmit={handleSubmit}
        className="home-action-form profile-card-form"
      >
        <label>
          Current password
          <input
            name="currentPassword"
            type="password"
            autoComplete="current-password"
            required
          />
        </label>
        <div className="home-form-grid">
          <label>
            New password
            <input
              name="newPassword"
              type="password"
              autoComplete="new-password"
              required
              minLength={8}
              maxLength={128}
            />
          </label>
          <label>
            Confirm new password
            <input
              name="confirmPassword"
              type="password"
              autoComplete="new-password"
              required
            />
          </label>
        </div>
        <div className="profile-status-region" role="status" aria-live="polite">
          {visibleError && (
            <p className={`home-action-message is-${visibleStatus}`}>{visibleError}</p>
          )}
        </div>
        <SubmitButton label="Change password" pendingLabel="Changing…" />
      </form>
    </section>
  );
}

function LogoutCard() {
  return (
    <section className="home-card" aria-label="Log out">
      <div className="profile-logout-zone">
        <div>
          <p className="profile-logout-label">Log out</p>
          <p className="profile-logout-sub">Sign out of OweNone on this device.</p>
        </div>
        <LogoutForm />
      </div>
    </section>
  );
}

function SubmitButton({ label, pendingLabel }: { label: string; pendingLabel: string }) {
  const { pending } = useFormStatus();
  return (
    <button type="submit" className="home-btn home-btn-primary" disabled={pending}>
      {pending ? pendingLabel : label}
    </button>
  );
}
