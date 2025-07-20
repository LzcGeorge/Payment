Page({
  data: {
    balance: 0,
    transferLogs: [],
    loading: false,
    openid: 'test_openid_001', // 测试用，实际要用 wx.login 拿
  },

  onLoad() {
    this.fetchBalance();
    this.fetchLogs();
  },

  // 红包签到
  onSignIn() {
    const that = this;
    that.setData({ loading: true });
    wx.request({
      url: 'http://wepay.selfknow.cn/transfer/to_user',
      method: 'POST',
      header: { 'content-type': 'application/json' },
      data: {
        openid: that.data.openid,
        amount: 100,  // 分
        remark: '红包签到'
      },
      success(resp) {
        wx.showToast({ title: resp.data.msg || '签到成功' });
        that.fetchBalance();
        that.fetchLogs();
      },
      complete() {
        that.setData({ loading: false });
      }
    });
  },

  fetchBalance() {
    const that = this;
    wx.request({
      url: 'http://localhost:8080/user/balance?openid=' + that.data.openid,
      success(res) {
        that.setData({ balance: res.data.balance || 0 });
      }
    });
  },

  fetchLogs() {
    const that = this;
    wx.request({
      url: 'http://localhost:8080/transfer/logs?openid=' + that.data.openid,
      success(res) {
        that.setData({ transferLogs: res.data.data || [] });
      }
    });
  }
})