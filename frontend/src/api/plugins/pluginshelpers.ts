export type Hook = { signature: string; type: string };

export type Hooks = Record<string, Hook>;

export type Plugin = {
    enabled: boolean;
    name: string;
    version: string;
    hooks: Hooks;
};
