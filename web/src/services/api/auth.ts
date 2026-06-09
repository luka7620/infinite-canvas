import { apiGet, apiPost } from "@/services/api/request";

export const AUTH_TOKEN_KEY = "infinite-canvas-auth-token-v1";

export type UserRole = "guest" | "user" | "admin";

export type AuthUser = {
    id: string;
    username: string;
    displayName: string;
    avatarUrl: string;
    role: UserRole;
    credits: number;
    lastCheckInDate: string;
    checkedInToday: boolean;
    createdAt: string;
    updatedAt: string;
};

export type AuthSession = {
    token: string;
    user: AuthUser;
};

export type CheckInResult = {
    user: AuthUser;
    credits: number;
};

export type InviteCodeRedeemResult = {
    user: AuthUser;
    inviteCode: {
        id: string;
        code: string;
        type: "register" | "credits";
        credits: number;
    };
    credits: number;
};

export type AuthPayload = {
    username: string;
    password: string;
    code?: string;
};

export type ProfilePayload = {
    username: string;
    displayName?: string;
    avatarUrl?: string;
};

export async function login(payload: AuthPayload) {
    return apiPost<AuthSession>("/api/auth/login", payload);
}

export async function register(payload: AuthPayload) {
    return apiPost<AuthSession>("/api/auth/register", payload);
}

export async function fetchCurrentUser(token?: string) {
    return apiGet<AuthUser>("/api/auth/me", undefined, token);
}

export async function saveProfile(token: string, payload: ProfilePayload) {
    return apiPost<AuthUser>("/api/auth/profile", payload, token);
}

export async function checkIn(token: string) {
    return apiPost<CheckInResult>("/api/auth/check-in", {}, token);
}

export async function redeemInviteCode(token: string, code: string) {
    return apiPost<InviteCodeRedeemResult>("/api/auth/invite-codes/redeem", { code }, token);
}
