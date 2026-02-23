export type Perm = { description: string; name: string; permission: string };
export type PermedUser = {
    user_id: number;
    permissions?: Perm[];
};

export function ExtractPermsList(
    user: PermedUser | undefined,
): string[] | undefined {
    return user?.permissions?.map((v) => v.permission);
}

export function UserHasPerm(
    user: PermedUser | undefined,
    permission: string,
): boolean {
    return ExtractPermsList(user)?.includes(permission) ?? false;
}

export function ParsePermsQuery(perms: string[]): string {
    return perms
        .map((perm) => `permission=${encodeURIComponent(perm)}`)
        .join("&");
}
