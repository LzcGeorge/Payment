Page({
  data: {
    balance: 0,
    transferLogs: [],
    loading: false,
    openid: 'test_openid_003', // 测试用，实际要用 wx.login 拿
    transferResult : null,
    out_bill_no: "",
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
        amount:  Math.floor(Math.random() * 49) + 1,  // 分
        remark: '红包签到'
      },
      success: (res) => {
        console.log(res);
        if (res.data && res.data.package_info) {
          wx.showToast({ title: res.data.msg || '签到成功', icon: 'success' });
          this.data.out_bill_no = res.data.out_bill_no;
          console.log(this.data.out_bill_no);
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

  // 确认转账（手动触发notify）
  onConfirmTransfer() {
    if (!this.data.out_bill_no) {
      wx.showToast({ title: '没有单号', icon: 'none' });
      return;
    }
    wx.request({
      url: 'http://wepay.selfknow.cn/transfer/confirm',
      method: 'POST',
      data: {
        out_bill_no: this.data.out_bill_no
      },
      success: (res) => {
        console.log(this.data.out_bill_no)
        if (res.statusCode === 200) {
          wx.showToast({ title: '转账已确认', icon: 'success' });
          // 这里可选择重新拉取余额、转账记录等
          this.data.out_bill_no = ""
          this.fetchBalance();
        } else {
          wx.showToast({ title: res.data.msg || '确认失败', icon: 'none' });
        }
      },
      fail: () => {
        wx.showToast({ title: '网络异常', icon: 'none' });
      }
    });
  },
  // 模拟 requestMerchantTransfer 行为
  startTransfer(payParams) {
      wx.request({
        url: 'http://wepay.selfknow.cn/transfer/notify',
        method: 'POST',
        header: { 'content-type': 'application/json' },
        package: 'affffddafdfafddffda==',
        data: {
          outbillno: payParams.out_bill_no
        },
        success: (res) => {
          if (res.statusCode === 200) {
            wx.showToast({ title: res.data.msg || '签到成功', icon: 'success' });
          } else {
            wx.showToast({ title: res.data.msg || '签到失败', icon: 'none' });
          }
        },
        fail: (res) => {
          console.log('fail:', res);
        },

      })
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