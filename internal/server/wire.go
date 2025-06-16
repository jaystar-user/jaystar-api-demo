//go:build wireinject
// +build wireinject

package server

import (
	logUtils "github.com/SeanZhenggg/go-utils/logger"
	"github.com/google/wire"
	"jaystar/internal/app/job"
	web "jaystar/internal/app/web"
	"jaystar/internal/config"
	jobCtrl "jaystar/internal/controller/job"
	jobMw "jaystar/internal/controller/job/middleware"
	webCtrl "jaystar/internal/controller/web"
	webMw "jaystar/internal/controller/web/middleware"
	"jaystar/internal/controller/web/util"
	"jaystar/internal/database"
	"jaystar/internal/interfaces"
	"jaystar/internal/repository"
	"jaystar/internal/service"
	commonSrv "jaystar/internal/service/common"
	"jaystar/internal/utils/kintoneAPI"
)

func NewAppServer() *appServer {
	panic(
		wire.Build(
			config.ProviderIConfigEnv,
			wire.Bind(new(logUtils.ILogConfig), new(config.IConfigEnv)),
			logUtils.ProviderILogger,
			kintoneAPI.ProvideKintoneClient,
			util.ProviderRequestParse,

			webMw.ProvideResponseMiddleware,
			webMw.ProvideHttpLogMiddleware,
			webMw.ProvideAuthMiddleware,
			webMw.ProvideRecoverMiddleware,
			webMw.ProvideInternalAuthMiddleware,

			jobMw.ProvideJobLogMiddleware,
			wire.Bind(new(jobMw.IJobMiddleware), new(*jobMw.JobLogMiddleware)),

			database.ProvidePostgresDB,

			repository.ProvideUserRepository,
			wire.Bind(new(interfaces.IUserRepo), new(*repository.UserRepo)),

			repository.ProvideStudentRepository,
			wire.Bind(new(interfaces.IStudentRepo), new(*repository.StudentRepo)),

			repository.ProvideKintoneStudentRepository,
			wire.Bind(new(interfaces.IKintoneStudentRepo), new(*repository.KintoneStudentRepository)),

			repository.ProvideDepositRecordRepository,
			wire.Bind(new(interfaces.IDepositRecordRepo), new(*repository.DepositRecordRepo)),

			repository.ProvideKintoneDepositRecordRepository,
			wire.Bind(new(interfaces.IKintoneDepositRecordRepo), new(*repository.KintoneDepositRecordRepository)),

			repository.ProvideReduceRecordRepository,
			wire.Bind(new(interfaces.IReduceRecordRepo), new(*repository.ReduceRecordRepo)),

			repository.ProvideKintoneReduceRecordRepository,
			wire.Bind(new(interfaces.IKintoneReduceRecordRepo), new(*repository.KintoneReduceRecordRepository)),

			repository.ProvideScheduleRepository,
			wire.Bind(new(interfaces.IScheduleRepo), new(*repository.ScheduleRepo)),

			repository.ProvideKintoneScheduleRepository,
			wire.Bind(new(interfaces.IKintoneScheduleRepo), new(*repository.KintoneScheduleRepository)),

			repository.ProvidePointCardRepository,
			wire.Bind(new(interfaces.IPointCardRepo), new(*repository.PointCardRepo)),

			repository.ProvideKintonePointCardRepository,
			wire.Bind(new(interfaces.IKintonePointCardRepo), new(*repository.KintonePointCardRepository)),

			repository.ProvideSemesterSettleRecordRepository,
			wire.Bind(new(interfaces.ISemesterSettleRecordRepo), new(*repository.SemesterSettleRecordRepository)),

			repository.ProvideKintoneSemesterSettleRecordRepository,
			wire.Bind(new(interfaces.IKintoneSemesterSettleRecordRepo), new(*repository.KintoneSemesterSettleRecordRepository)),

			commonSrv.ProvideUserCommonService,
			wire.Bind(new(interfaces.IUserCommonSrv), new(*commonSrv.UserCommonService)),

			commonSrv.ProvideStudentCommonService,
			wire.Bind(new(interfaces.IStudentCommonSrv), new(*commonSrv.StudentCommonService)),

			service.ProvidePointCardService,
			wire.Bind(new(interfaces.IPointCardSrv), new(*service.PointCardService)),

			commonSrv.ProvideDepositRecordCommonService,
			wire.Bind(new(interfaces.IDepositRecordCommonSrv), new(*commonSrv.DepositRecordCommonService)),

			commonSrv.ProvideReduceRecordCommonService,
			wire.Bind(new(interfaces.IReduceRecordCommonSrv), new(*commonSrv.ReduceRecordCommonService)),

			commonSrv.ProvideScheduleCommonService,
			wire.Bind(new(interfaces.IScheduleCommonSrv), new(*commonSrv.ScheduleCommonService)),

			service.ProvideUserService,
			wire.Bind(new(interfaces.IUserSrv), new(*service.UserService)),

			service.ProvideStudentService,
			wire.Bind(new(interfaces.IStudentSrv), new(*service.StudentService)),

			service.ProvideDepositRecordService,
			wire.Bind(new(interfaces.IDepositRecordSrv), new(*service.DepositRecordService)),

			service.ProvideReduceRecordService,
			wire.Bind(new(interfaces.IReduceRecordSrv), new(*service.ReduceRecordService)),

			service.ProvideScheduleService,
			wire.Bind(new(interfaces.IScheduleSrv), new(*service.ScheduleService)),

			service.ProvideSemesterSettleRecordService,
			wire.Bind(new(interfaces.ISemesterSettleRecordSrv), new(*service.SemesterSettleRecordService)),

			webCtrl.ProvideUserController,

			webCtrl.ProvideStudentController,

			webCtrl.ProvideScheduleController,

			webCtrl.ProvideDepositRecordController,

			webCtrl.ProvideReduceRecordController,

			webCtrl.ProvideSemesterSettleRecordController,

			webCtrl.ProvidePointCardController,

			webCtrl.ProvideSyncController,

			webCtrl.ProvideController,

			jobCtrl.ProvideController,

			web.ProvideWebApp,

			job.ProvideJob,

			wire.Struct(new(appServer), "*"),
		),
	)
}
