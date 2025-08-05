export const baseUrlApi = (url: string) =>
    process.env.NODE_ENV === "development"
        ? `/api/${url}`
        : `http://127.0.0.1:3000/${url}`;
