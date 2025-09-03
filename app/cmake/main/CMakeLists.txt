set(APP_SRCS
        ../../src/main/main.cpp
)
if (OS_MACOSX OR OS_WINDOWS)
    set(APP_RESOURCES_SRCS ${SHARED_RESOURCES_SRCS} ${SHARED_ASSETS_DIR})
endif ()
set(APP_TARGET "CefLib")
if (OS_MAC)
    set(APP_HELPER_TARGET "MyApp_Helper")
    set(APP_HELPER_OUTPUT_NAME "MyApp Helper")
endif ()
include_directories(../../include)
if (OS_WINDOWS)
    set(APP_SRCS ${APP_SRCS} ${APP_RESOURCES_SRCS})
    set_app_source_groups("${APP_SRCS}")
    add_executable(${APP_TARGET} WIN32 ${APP_SRCS})
    set_app_executable_target_properties(${APP_TARGET})
    add_dependencies(${APP_TARGET} shared)
    target_link_libraries(${APP_TARGET} shared)
#    add_custom_command(
#            TARGET ${APP_TARGET}
#            POST_BUILD
#            COMMAND
#            ${CMAKE_COMMAND} -E copy_directory
#            "${CMAKE_CURRENT_SOURCE_DIR}/${SHARED_ASSETS_SRCS}"
#            VERBATIM)
endif ()
