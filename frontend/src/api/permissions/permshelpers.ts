export type Perm = { description: string; name: string; permission: string };
export type PermedUser = {
    user_id: number;
    permissions?: Perm[];
};

export function ExtractPermsList(user: PermedUser): string[] | undefined {
    return user.permissions?.map((v) => v.permission);
}

export function UserHasPerm(
    user: PermedUser | undefined,
    permission: string,
): boolean {
    if (!user) {
        return false;
    }
    return ExtractPermsList(user)?.includes(permission) ?? false;
}

export function ParsePermsQuery(perms: string[]): string {
    return perms
        .map((perm) => `permission=${encodeURIComponent(perm)}`)
        .join("&");
}
