// ==UserScript==
// @name         小红书资源保存器
// @namespace    http://tampermonkey.net/
// @version      1.0
// @description  将小红书页面资源保存到网盘
// @author       You
// @match        https://www.xiaohongshu.com/*
// @icon         https://www.google.com/s2/favicons?sz=64&domain=xiaohongshu.com
// @grant        GM_addStyle
// @grant        GM_xmlhttpRequest
// ==/UserScript==

(function () {
    'use strict';

    // 配置常量
    // API接口地址 - 用于发送小红书资源链接到后端处理
    // 修改此变量可以切换不同的后端服务地址
    const API_ENDPOINT = 'http://example.com:9092/api/process';

    // 添加样式
    GM_addStyle(`
        .save-to-cloud-btn {
            position: fixed;
            top: 20px;
            right: 20px;
            background-color: #4CAF50;
            color: white;
            border: none;
            padding: 12px 24px;
            text-align: center;
            text-decoration: none;
            display: inline-block;
            font-size: 16px;
            cursor: pointer;
            border-radius: 5px;
            transition: background-color 0.3s;
            z-index: 9999;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
        }
        .save-to-cloud-btn:hover {
            background-color: #45a049;
        }
        .save-to-cloud-btn:disabled {
            background-color: #cccccc;
            cursor: not-allowed;
        }
        .save-status-modal {
            position: fixed;
            top: 20px;
            left: 50%;
            transform: translateX(-50%);
            background-color: white;
            padding: 25px;
            border-radius: 10px;
            box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
            z-index: 10000;
            min-width: 300px;
            text-align: center;
            animation: fadeInOut 3s ease-in-out forwards;
        }
        .save-status-modal h3 {
            margin-top: 0;
            margin-bottom: 15px;
            font-size: 18px;
        }
        .save-status-modal.success {
            border-left: 4px solid #4CAF50;
        }
        .save-status-modal.error {
            border-left: 4px solid #F44336;
        }
        .save-status-modal .message {
            margin-bottom: 20px;
            line-height: 1.5;
        }
        
        @keyframes fadeInOut {
            0% { opacity: 0; transform: translate(-50%, -20px); }
            20% { opacity: 1; transform: translate(-50%, 0); }
            80% { opacity: 1; transform: translate(-50%, 0); }
            100% { opacity: 0; transform: translate(-50%, -20px); }
        }
        .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: rgba(0, 0, 0, 0.5);
            z-index: 9999;
        }
    `);

    // 创建保存按钮
    function createSaveButton() {
        const button = document.createElement('button');
        button.className = 'save-to-cloud-btn';
        button.textContent = '保存到网盘';

        button.addEventListener('click', async function () {
            this.disabled = true;
            this.textContent = '保存中...';

            try {
                const currentUrl = window.location.href;
                await saveToCloud(currentUrl);
            } catch (error) {
                showStatusModal('保存失败', error.message, 'error');
            } finally {
                this.disabled = false;
                this.textContent = '保存到网盘';
            }
        });

        document.body.appendChild(button);
    }

    // 保存到网盘
    function saveToCloud(url) {
        return new Promise((resolve, reject) => {
            // 构造请求数据
            const requestData = JSON.stringify({
                message: {
                    text: url
                }
            });

            // 发送请求到后端API
            GM_xmlhttpRequest({
                method: 'POST',
                url: API_ENDPOINT, // 使用配置的API接口地址
                headers: {
                    'Content-Type': 'application/json'
                },
                data: requestData,
                onload: function (response) {
                    if (response.status >= 200 && response.status < 300) {
                        try {
                            const data = JSON.parse(response.responseText);
                            showStatusModal('保存成功', '资源已成功提交到服务器', 'success');
                            resolve(data);
                        } catch (e) {
                            reject(new Error('解析服务器响应失败'));
                        }
                    } else {
                        reject(new Error(`服务器响应错误: ${response.status}`));
                    }
                },
                onerror: function () {
                    reject(new Error('网络请求失败，请检查服务器是否运行'));
                }
            });
        });
    }

    // 显示状态弹窗（自动消失）
    function showStatusModal(title, message, type) {
        // 创建弹窗
        const modal = document.createElement('div');
        modal.className = `save-status-modal ${type}`;

        modal.innerHTML = `
            <h3>${title}</h3>
            <div class="message">${message}</div>
        `;

        // 添加到页面
        document.body.appendChild(modal);

        // 3秒后自动移除弹窗
        setTimeout(() => {
            if (document.body.contains(modal)) {
                document.body.removeChild(modal);
            }
        }, 3000);
    }

    // 页面加载完成后执行
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', createSaveButton);
    } else {
        createSaveButton();
    }
})();