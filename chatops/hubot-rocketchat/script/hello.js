module.exports=(robot) => {
robot.hear(/hello/i,(res) => {
const userName = res.message.user.name;
res.reply(`Hello,${userName}!`);
});
};
