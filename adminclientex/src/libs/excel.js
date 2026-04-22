import XLSX from 'xlsx'

export function exportExcel(columns, data, filename = '导出数据') {
  // 处理表头
  const headers = columns.map(item => {
    return {
      header: item.title,
      key: item.key,
      width: item.width / 5 || 20 // 调整列宽
    }
  })
  
  // 创建工作簿
  const wb = XLSX.utils.book_new()
  
  // 处理数据
  const wsData = [
    headers.map(item => item.header), // 表头行
    ...data.map(item => {
      return headers.map(header => {
        return item[header.key]
      })
    })
  ]
  
  // 创建工作表
  const ws = XLSX.utils.aoa_to_sheet(wsData)
  
  // 设置列宽
  ws['!cols'] = headers.map(item => ({ wch: item.width }))
  
  // 将工作表添加到工作簿
  XLSX.utils.book_append_sheet(wb, ws, 'Sheet1')
  
  // 生成Excel文件并下载
  const wbout = XLSX.write(wb, { bookType: 'xlsx', type: 'array' })

  return wbout
}

export default {
    exportExcel,
}
