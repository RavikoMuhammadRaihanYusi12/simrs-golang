"use client";

// ponytail: client-side auth only. Token in localStorage, role from JWT payload
// (unverified — server enforces real authz). Upgrade: httpOnly cookie + backend guard.
const TOKEN_KEY = "simrs_token";
const ROLE_KEY = "simrs_role";

export function saveAuth(token: string) {
    localStorage.setItem(TOKEN_KEY, token);
    const payload = JSON.parse(atob(token.split(".")[1] || "e30="));
    localStorage.setItem(ROLE_KEY, payload.role || "");
    return payload.role as string;
}

export function getToken(): string | null {
    return typeof window === "undefined" ? null : localStorage.getItem(TOKEN_KEY);
}

export function getRole(): string | null {
    return typeof window === "undefined" ? null : localStorage.getItem(ROLE_KEY);
}

export function logout() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(ROLE_KEY);
}

export function requireAuth(): boolean {
    if (!getToken()) {
        window.location.href = "/login";
        return false;
    }
    return true;
}

export function homeForRole(role: string | null): string {
    return role === "admin" ? "/dashboard" : "/pasien";
}

// authFetch wraps fetch with the Authorization header. Redirects to /login on 401.
export async function authFetch(input: string, init: RequestInit = {}): Promise<Response> {
    const token = getToken();
    const headers = new Headers(init.headers);
    if (token) headers.set("Authorization", `Bearer ${token}`);
    const res = await fetch(input, { ...init, headers });
    if (res.status === 401) {
        logout();
        window.location.href = "/login";
    }
    return res;
}
