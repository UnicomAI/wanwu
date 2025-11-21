# Call test
import json
import requests
import time

BASE_URL = "http://localhost:20041"

if __name__ == "__main__":
    # ******************* Test China Unicom *******************
    url = f'{BASE_URL}/es/test'
    headers = {'Content-Type': 'application/json'}
    test = {"test": "test"}
    r = requests.post(url, data=json.dumps(test), headers=headers)
    print(r.status_code)
    result_data = json.loads(r.text)
    print(result_data)

    # ******************* rag/kn/init_kb *******************
    url = f"{BASE_URL}/rag/kn/init_kb"
    headers = {'Content-Type': 'application/json'}
    user_id = "hhh20240815"
    kb_name = "test_KB"
    print(f"url:{url},{user_id},kb_name:{kb_name}=============== init_kb start =================")
    # Construct request data
    data = {
        "userId": user_id,
        "kb_name": kb_name,
    }
    # Send POST request
    response = requests.post(url, headers=headers, json=data)  # Send JSON data directly using json parameters
    print(response.status_code)
    result_data = json.loads(response.text)
    print(result_data)
    print(f"=============== init_kb end =================")
    # ******************* rag/kn/init_kb *******************

    # ******************* rag/kn/add ****************************
    url = f"{BASE_URL}/rag/kn/add"
    headers = {'Content-Type': 'application/json'}

    user_id = "hhh20240815"
    kb_name = "test_KB"
    print(f"url:{url},{user_id},kb_name:{kb_name}===============  rag/kn/add start =================")
    data = [
        {
            'content': '中国联通网络和信息安全监控指挥中心  \n运营手册（终端安全分册）  \n（版本号： V1.0）  \n一、总则  \n1.1编制说明  \n为加强规范终端安全运营工作流程，明确运营场景的边界，明确运营场景人员职责，梳理现有各流程之间的逻辑关系。  \n1.2适用对象  \n本手册适用于中国联通集团SOC团队、终端安全运营团队、分指挥中心。适用于中国联通终端安全运营支撑场景。在集团领导小组的指挥下，终端安全保障和事件处置坚持统一指挥、分级负责、保障有力、减小事件的影响范围。  \n1.3流程与对象  \n终端安全运营流程：  \n终端安全运营对象：  \n|名称|职责|\n|集团SOC一线|完成SOC领导终端工作、大屏数据监控、定期开展勒索与挖矿病毒治理工作、数据汇总与报告|\n|集团SOC二线（终端团队）|提供高级终端安全技术支撑服务，联动厂商承接用户产品需求，解决产品功能问题，做好重大终端安全事件与重保期间终端安全事件应急响应；协助各省分子公司开展集团及本单位终端安全考核合规相关工作；|\n|分指挥中心|开展终端安全项目管理、终端病毒/漏洞事件处置、日常问题处置等相关工作，及时响应集团总指挥中心下达的指令工单，做好终端安全运营保障支撑；|\n|集团SOC二线（研判团队）|研判安全事件、闭环指令工单|\n二、终端安全运营内容  \n2.1终端入网  \n禁止高风险终端通过直连或远程的方式进入办公网。  \n2.1.2终端准入  \n已开启终端准入策略，提供入网设备的访问控制、入网安全合规性能力，保障全国联通入网终端可控性与合规性，并与VPN侧联动限制远程办公终端未安装终端安全管理系统客户端就无法使用VPN软件。  \n2.2终端基线薄弱  \n终端安全基线是终端安全的短板或者说是最基础的安全要求，一台没有经过任何安全基线加固的终端将成为不法分子首要攻击的目标。  \n2.2.1终端基线加固  \n终端安全基线加固是我们终端安全的基石，为奠定这一基石现已启用终端安全系统中的“运维管控”与“安全检查”功能模块便捷式的帮助终端使用人配置终端安全基线。  \n2.3终端感染病毒',
            'embedding_content': '中国联通网络和信息安全监控指挥中心  \n运营手册（终端安全分册）  \n（版本号： V1.0）  \n一、总则  \n1.1编制说明  \n为加强规范终端安全运营工作流程，明确运营场景的边界，明确运营场景人员职责，梳理现有各流程之间的逻辑关系。  \n1.2适用对象  \n本手册适用于中国联通集团SOC团队、终端安全运营团队、分指挥中心。适用于中国联通终端安全运营支撑场景。在集团领导小组的指挥下，终端安全保障和事件处置坚持统一指挥、分级负责、保障有力、减小事件的影响范围。  \n1.3流程与对象  \n终端安全运营流程：  \n终端安全运营对象：  \n|名称|职责|\n|集团SOC一线|完成SOC领导终端工作、大屏数据监控、定期开展勒索与挖矿病毒治理工作、数据汇总与报告|\n|集团SOC二线（终端团队）|提供高级终端安全技术支撑服务，联动厂商承接用户产品需求，解决产品功能问题，做好重大终端安全事件与重保期间终端安全事件应急响应；协助各省分子公司开展集团及本单位终端安全考核合规相关工作；|\n|分指挥中心|开展终端安全项目管理、终端病毒/漏洞事件处置、日常问题处置等相关工作，及时响应集团总指挥中心下达的指令工单，做好终端安全运营保障支撑；|\n|集团SOC二线（研判团队）|研判安全事件、闭环指令工单|\n二、终端安全运营内容  \n2.1终端入网  \n禁止高风险终端通过直连或远程的方式进入办公网。  \n2.1.2终端准入  \n已开启终端准入策略，提供入网设备的访问控制、入网安全合规性能力，保障全国联通入网终端可控性与合规性，并与VPN侧联动限制远程办公终端未安装终端安全管理系统客户端就无法使用VPN软件。  \n2.2终端基线薄弱  \n终端安全基线是终端安全的短板或者说是最基础的安全要求，一台没有经过任何安全基线加固的终端将成为不法分子首要攻击的目标。  \n2.2.1终端基线加固  \n终端安全基线加固是我们终端安全的基石，为奠定这一基石现已启用终端安全系统中的“运维管控”与“安全检查”功能模块便捷式的帮助终端使用人配置终端安全基线。  \n2.3终端感染病毒',
            'meta_data': {'file_name': '中国联通网络和信息安全监控指挥中心运营手册（终端安全分册）.docx'}},
        {
            'content': '终端安全基线是终端安全的短板或者说是最基础的安全要求，一台没有经过任何安全基线加固的终端将成为不法分子首要攻击的目标。  \n2.2.1终端基线加固  \n终端安全基线加固是我们终端安全的基石，为奠定这一基石现已启用终端安全系统中的“运维管控”与“安全检查”功能模块便捷式的帮助终端使用人配置终端安全基线。  \n2.3终端感染病毒  \n终端病毒是人为制造的，有破坏性，又有传染性和潜伏性的，对终端信息或系统起破坏作用的程序。  \n2.3.1终端病毒扫描  \n病毒扫描的工作旨在为终端清理已下载的病毒文件以及防护将下载与打开的病毒文件或程序。为了更加及时的处置全国高风险染毒终端，现采用终端安全可视化系统进行全国高风险终端事件监控，对于感染勒索或挖矿病毒且未处理的终端通过统一指挥调度平台与终端安全邮箱双向通报，对于病毒查杀超500次未处理的终端实行统一指挥调度平台单向通报。对于病毒防护方面全国策略均已开启“云查杀”防护引擎、文件实时防护、主动防御等功能，极大的提高了终端防护病毒的能力。  \n2.3.2数据输出  \n1、日报中告警终端事件数与告警终端事件数。  \n2、周报中勒索挖矿病毒指令下发数、挖矿病毒事件数、勒索病毒事件数、高风险病毒终端指令下发数、终端指令闭环率、高风险病毒终端事件数与感染终端省份分布情况等。  \n3、月报中指令下发数、挖矿/勒索病毒省份分布情况、终端指令闭环情况与高风险病毒终端省份分布情况。  \n2.4终端漏洞暴露  \n计算机漏洞是在硬件、软件、协议的具体实现或系统安全策略上存在的缺陷，从而可以使攻击者能够在未授权的情况下访问或破坏系统。  \n2.4.1终端漏洞修复  \n终端漏洞修复工作涉及终端较多且漏洞修复工作可能出现较多不稳定因素，目前个别省份开启漏洞自动修复功能，仅集团对特定高危漏洞进行后台强制修复，对于其他漏洞则由终端使用人根据自身情况修复。  \n2.5终端安全可视化系统运营',
            'embedding_content': '终端安全基线是终端安全的短板或者说是最基础的安全要求，一台没有经过任何安全基线加固的终端将成为不法分子首要攻击的目标。  \n2.2.1终端基线加固  \n终端安全基线加固是我们终端安全的基石，为奠定这一基石现已启用终端安全系统中的“运维管控”与“安全检查”功能模块便捷式的帮助终端使用人配置终端安全基线。  \n2.3终端感染病毒  \n终端病毒是人为制造的，有破坏性，又有传染性和潜伏性的，对终端信息或系统起破坏作用的程序。  \n2.3.1终端病毒扫描  \n病毒扫描的工作旨在为终端清理已下载的病毒文件以及防护将下载与打开的病毒文件或程序。为了更加及时的处置全国高风险染毒终端，现采用终端安全可视化系统进行全国高风险终端事件监控，对于感染勒索或挖矿病毒且未处理的终端通过统一指挥调度平台与终端安全邮箱双向通报，对于病毒查杀超500次未处理的终端实行统一指挥调度平台单向通报。对于病毒防护方面全国策略均已开启“云查杀”防护引擎、文件实时防护、主动防御等功能，极大的提高了终端防护病毒的能力。  \n2.3.2数据输出  \n1、日报中告警终端事件数与告警终端事件数。  \n2、周报中勒索挖矿病毒指令下发数、挖矿病毒事件数、勒索病毒事件数、高风险病毒终端指令下发数、终端指令闭环率、高风险病毒终端事件数与感染终端省份分布情况等。  \n3、月报中指令下发数、挖矿/勒索病毒省份分布情况、终端指令闭环情况与高风险病毒终端省份分布情况。  \n2.4终端漏洞暴露  \n计算机漏洞是在硬件、软件、协议的具体实现或系统安全策略上存在的缺陷，从而可以使攻击者能够在未授权的情况下访问或破坏系统。  \n2.4.1终端漏洞修复  \n终端漏洞修复工作涉及终端较多且漏洞修复工作可能出现较多不稳定因素，目前个别省份开启漏洞自动修复功能，仅集团对特定高危漏洞进行后台强制修复，对于其他漏洞则由终端使用人根据自身情况修复。  \n2.5终端安全可视化系统运营',
            'meta_data': {'file_name': '中国联通网络和信息安全监控指挥中心运营手册（终端安全分册）.docx'}},
    ]
    # Construct request data
    data = {
        "userId": user_id,
        "kb_name": kb_name,
        "data": data
    }
    # Send POST request
    response = requests.post(url, headers=headers, json=data)  # Send JSON data directly using json parameters
    print(response.status_code)
    result_data = json.loads(response.text)
    print(result_data)
    print(f"=============== rag/kn/add end =================")

    # ************** rag/kn/list_kb_names *****************
    url = f"{BASE_URL}/rag/kn/list_kb_names"
    headers = {'Content-Type': 'application/json'}

    user_id = "hhh20240815"
    print(f"url:{url},{user_id}, =============== /rag/kn/list_kb_names start =================")
    # Construct request data
    data = {
        "userId": user_id,
    }

    # Send POST request
    response = requests.post(url, headers=headers, json=data)  # Send JSON data directly using json parameters
    print(response.status_code)
    result_data = json.loads(response.text)
    print(result_data)

    # ************** rag/kn/list_file_names *****************
    url = f"{BASE_URL}/rag/kn/list_file_names"
    headers = {'Content-Type': 'application/json'}

    user_id = "hhh20240815"
    kb_name = "test_KB"
    print(f"url:{url},{user_id},kb_name:{kb_name}===============  rag/kn/list_file_names start =================")

    # Construct request data
    data = {
        "userId": user_id,
        "kb_name": kb_name,
    }
    # Send POST request
    response = requests.post(url, headers=headers, json=data)  # Send JSON data directly using json parameters
    print(response.status_code)
    result_data = json.loads(response.text)
    print(result_data)

    # ************** rag/kn/search *****************
    url = f"{BASE_URL}/rag/kn/search"
    headers = {'Content-Type': 'application/json'}

    user_id = "hhh20240815"
    kb_names = ['test_KB']
    question = "连续执政最长的政党"
    top_k = 10
    min_score = 0
    print(f"url:{url},{user_id},kb_names:{kb_names}===============  /rag/kn/search start =================")
    # Construct request data
    data = {
        "userId": user_id,
        "kb_names": kb_names,
        "question": question,
        "topk": top_k,
        "threshold": min_score
    }
    # Send POST request
    response = requests.post(url, headers=headers, json=data)  # Send JSON data directly using json parameters
    print(response.status_code)
    result_data = json.loads(response.text)
    print(result_data)

    # ************** rag/kn/del_kb *****************
    url = f"{BASE_URL}/rag/kn/del_kb"
    headers = {'Content-Type': 'application/json'}

    user_id = "hhh20240815"
    kb_name = "test_KB"
    print(f"url:{url},{user_id},kb_names:{kb_names}===============  /rag/kn/del_kb start =================")
    # Construct request data
    data = {
        "userId": user_id,
        "kb_name": kb_name,
    }

    # Send POST request
    response = requests.post(url, headers=headers, json=data)  # Send JSON data directly using json parameters
    print(response.status_code)
    result_data = json.loads(response.text)
    print(result_data)
