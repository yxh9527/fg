import dayjs from "dayjs";

export const safeNumber = (value) => {
  if (value === null || value === undefined || value === "") return 0;
  return Number(value) || 0;
};

export const toFixedValue = (value, digits = 2) => safeNumber(value).toFixed(digits);

export const formatDateTime = (value) => {
  if (!value) return "";
  return dayjs(value * 1000).format("YYYY-MM-DD HH:mm:ss");
};

export const formatPickerDayStart = (value) =>
  dayjs(Number(value)).startOf("day").valueOf();

export const formatPickerDayEnd = (value) =>
  dayjs(Number(value)).endOf("day").valueOf();
