/** A wrapper for Set<T> that is designed for use with useEffect/useState and is immutable */
export default class FreezeTagSet<T> {
    private set: Set<T>;

    constructor(iterable?: Iterable<T>) {
        this.set = new Set(iterable);
    }

    private copy(set: Set<T>): Set<T> {
        return new Set(set.values());
    }

    private setToArray(set: Set<T>): T[] {
        return set.values().toArray();
    }

    add(...items: T[]): FreezeTagSet<T> {
        const arr = this.setToArray(this.set);
        arr.push(...items);
        return new FreezeTagSet(arr);
    }

    remove(...items: T[]): FreezeTagSet<T> {
        const copy = this.copy(this.set);
        for (const item of items) {
            copy.delete(item);
        }
        return new FreezeTagSet(copy.values());
    }

    /** Returns true if all items are in the set */
    has(...items: T[]): boolean {
        for (const item of items) {
            if (this.set.has(item)) {
                continue;
            } else {
                return false;
            }
        }
        return true;
    }

    /** Returns a new set where if each item was present in the old set, it is removed and if it wasn't, it is added */
    toggle(...items: T[]): FreezeTagSet<T> {
        const copy = this.copy(this.set);
        for (const item of items) {
            if (copy.has(item)) {
                copy.delete(item);
            } else {
                copy.add(item);
            }
        }
        return new FreezeTagSet(copy.values());
    }

    size(): number {
        return this.set.size;
    }

    toArray(): T[] {
        return this.setToArray(this.set);
    }

    clone(): FreezeTagSet<T> {
        return new FreezeTagSet(this.toArray());
    }

    clear(): FreezeTagSet<T> {
        return new FreezeTagSet();
    }
}
