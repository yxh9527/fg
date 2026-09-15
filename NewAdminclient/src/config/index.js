const config = {
  title: "游戏管理后台",
  baseUrl: {
    // dev: "http://172.21.211.219:9529/api/auth/",
    // pro: "http://172.21.211.219:9529/api/auth/",
    dev: "/api/auth/",
    pro: "/api/auth/",
  },
  homeName: "new-home",
};

export const setting = {
  page: 1,
  pageSize: 15,
  pageOpts: [15, 30, 50, 100, 200, 300],
};

export default config;
