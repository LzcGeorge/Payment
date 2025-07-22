Page({
  data: {
    balance: 0,
    transferLogs: [],
    loading: false,
    openid: 'test_openid_003', // 测试用，实际要用 wx.login 拿
    appid: 'wxb9f4f763e5d4a6de',
    mchid: '1368139500',
    transferResult : null,
    package_info: "",
    out_bill_no: "",
  },

  onLoad() {
    this.fetchBalance();
  },

  // 红包签到
  onSignIn() {
    const that = this;
    that.setData({ loading: true });
    const pad = n => n < 10 ? '0' + n : n;
    const now = new Date();
    const time = 
      now.getFullYear() +
      pad(now.getMonth() + 1) +
      pad(now.getDate()) +
      pad(now.getHours()) +
      pad(now.getMinutes()) +
      pad(now.getSeconds());
    wx.request({
      url: 'http://wepay.selfknow.cn/transfer/to_user',
      method: 'POST',
      header: { 'content-type': 'application/json' },
      data: {
        openid: that.data.openid,
        amount:  Math.floor(Math.random() * 49) + 1,  // 分
        remark: '红包签到',
        time: time,
      },
      success: (res) => {
        console.log(res);
        if (res.data && res.data.package_info) {
          wx.showToast({ title: res.data.msg || '签到成功', icon: 'success' });
          this.setData({package_info: res.data.package_info})
          console.log(res.data.out_bill_no);
        } else {
          wx.showToast({ title: res.data.msg || '签到失败', icon: 'none' });
        }
      },
      fail: (err) => {
        wx.showToast({ title: '网络错误', icon: 'none' });
      },
      complete: () => {
        this.setData({ loading: false });
      }
    });
  },

  // 确认转账： 模拟 requestMerchantTransfer 行为
  onConfirmTransfer() {
    if (!this.data.package_info) {
      wx.showToast({ title: '没有红包', icon: 'none' });
      return;
    }

    wx.request({
      url: 'http://wepay.selfknow.cn/transfer/confirm',
      method: 'POST',
      data: {
        appid: this.data.appid,
        mch_id: this.data.mchid,
        package_info: this.data.package_info,
      },
      success: (res) => {
        if (res.statusCode === 200) {
          wx.showToast({ title: '转账已确认', icon: 'success' });
          // 这里可选择重新拉取余额、转账记录等
          this.fetchBalance();
          this.setData({package_info: ""})
          
        } else {
          wx.showToast({ title: res.data.msg || '微信平台还没处理完转账（没有 notify）', icon: 'none' });
        }
      },
      fail: () => {
        wx.showToast({ title: '网络异常', icon: 'none' });
      }
    });
  },
  
  

  fetchBalance() {
    const that = this;
    wx.request({
      url: 'http://wepay.selfknow.cn/transfer/amount?openid=' + that.data.openid,
      method: 'GET',
      header: { 'content-type': 'application/json' },
      success(res) {
        console.log(res);
        that.setData({ balance: res.data || 0});
      }
    });
  },

  
})