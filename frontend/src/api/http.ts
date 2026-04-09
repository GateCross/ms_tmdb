import axios from "axios";

const http = axios.create({
  baseURL: "/",
  timeout: 15000,
});

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const data = error?.response?.data;
    const msg =
      (typeof data === "string" ? data : "") ||
      data?.status_message ||
      data?.error ||
      data?.message ||
      data?.msg ||
      error?.message ||
      "请求失败";
    return Promise.reject(new Error(msg));
  },
);

export default http;
