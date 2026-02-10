export default async function mapWithConcurrency<T, R>(
    items: T[],
    limit: number,
    fn: (item: T, index: number) => Promise<R>,
): Promise<R[]> {
    const results = new Array<R>(items.length);
    let next = 0;

    // Worker function
    const worker = async () => {
        while (true) {
            const i = next++;
            if (i >= items.length) return;
            results[i] = await fn(items[i], i);
        }
    };

    // Start workers
    await Promise.all(
        Array.from({ length: Math.min(limit, items.length) }, () => worker()),
    );

    return results;
}
